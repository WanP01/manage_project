package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// Register for grpc server
// 服务端注册信息结构体
type Register struct {
	EtcdAddrs   []string // etcd 节点
	DialTimeout int      // etcd 连接超时时间

	closeCh     chan struct{}                           // 关闭通道
	leasesID    clientv3.LeaseID                        // 租约ID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse // 心跳通道

	srvInfo Server           // grpc 服务器信息
	srvTTL  int64            // grpc 服务ttl
	cli     *clientv3.Client // grpc 新建对Etcd 的连接
	logger  *zap.Logger      // 日志
}

// NewRegister create a register base on etcd
func NewRegister(etcdAddrs []string, logger *zap.Logger) *Register {
	return &Register{
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Register 注册grpc 服务
// 1. 新建etcd 连接
// 2. 配置 注册信息 （Register 结构体）
// 3.
func (r *Register) Register(srvInfo Server, ttl int64) (chan<- struct{}, error) {

	var err error

	// 拆分“127.0.0.1:xxxx", 无IP地址则报错
	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return nil, errors.New("invalid ip")
	}
	//建立etcd连接（config 为配置项）
	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,                                // etcd 节点
		DialTimeout: time.Duration(r.DialTimeout) * time.Second, // 连接过期时间
	}); err != nil {
		return nil, err
	}

	// Register信息 填充 server (grpc 服务信息) && 该服务 ttl Lease 租约时间信息
	r.srvInfo = srvInfo
	r.srvTTL = ttl

	// 新建租约，并将租约与 key-value 绑定放入 etcd ，配置 r.keepAliveCh
	if err = r.register(); err != nil {
		return nil, err
	}

	r.closeCh = make(chan struct{})

	// grpc端 监控与etcd的保活通道，失连则重新注册，同时支持grpc端关闭连接
	go r.keepAlive()

	return r.closeCh, nil
}

// Stop Register 注册grpc 的关闭信号
func (r *Register) Stop() {
	r.closeCh <- struct{}{} // 向 r.closeCh 通道发送信号，触发关闭动作
}

// register 上传grpc信息
func (r *Register) register() error {

	// 新建租约的ctx（上下文时间限制 r.DialTimeout 为连接过期限时）
	leaseCtx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	// 获得对应ttl 时间的租约
	leaseResp, err := r.cli.Grant(leaseCtx, r.srvTTL)
	if err != nil {
		return err
	}

	// 租约ID
	r.leasesID = leaseResp.ID
	// r.keepAliveCh : <-chan *clientv3.LeaseKeepAliveResponse 该channel持续性回复response信息
	if r.keepAliveCh, err = r.cli.KeepAlive(context.Background(), leaseResp.ID); err != nil {
		return err
	}

	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}

	// etcd 存储 key:value ,并绑定租约 Lease
	// key 为 grpc 服务的 "/name/version/addr" , value 为grpc 服务的json化信息 Server {name,version,addr,weight}
	_, err = r.cli.Put(context.Background(), BuildRegPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
	return err
}

// unregister 删除对应grpc信息节点
func (r *Register) unregister() error {
	_, err := r.cli.Delete(context.Background(), BuildRegPath(r.srvInfo))
	return err
}

// keepAlive
func (r *Register) keepAlive() {
	ticker := time.NewTicker(time.Duration(r.srvTTL) * time.Second)
	for {
		select {
		case <-r.closeCh: // 关闭 grpc 与 etcd 的连接（心跳检测）
			if err := r.unregister(); err != nil { // 删除etcd上grpc 信息节点
				r.logger.Error("unregister failed", zap.Error(err))
			}
			if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil { // 撤销相关租约 以及绑定在改租约上的所有key
				r.logger.Error("revoke failed", zap.Error(err))
			}
			return // 跳出循环
		case res := <-r.keepAliveCh: // 收到心跳信息回复,确认通道是否关闭 （可能的场景：因为网络原因ttl过期）
			if res == nil { // 通道关闭的情况下会读取buffer信息或nil，res== nil意味着通道关闭
				if err := r.register(); err != nil { // 重新注册服务
					r.logger.Error("register failed", zap.Error(err))
				}
			}
		case <-ticker.C: // 定时检查是否存在 保活通道
			if r.keepAliveCh == nil { // r.keepAliveCh 不存在 则重新注册该grpc服务
				if err := r.register(); err != nil {
					r.logger.Error("register failed", zap.Error(err))
				}
			}
		}
	}
}

// UpdateHandler return http handler
// 用于更新etcd 服务内部的 weight信息
func (r *Register) UpdateHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 获取URL信息中需要更新的weight最新值
		wi := req.URL.Query().Get("weight")
		weight, err := strconv.Atoi(wi)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		//存入grpc Server 信息中并解析成要存储进 etcd 的json 格式串
		var update = func() error {
			r.srvInfo.Weight = int64(weight)
			data, err := json.Marshal(r.srvInfo)
			if err != nil {
				return err
			}
			// 更新etcd 中的grpc 对应节点信息
			_, err = r.cli.Put(context.Background(), BuildRegPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
			return err
		}

		if err := update(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte("update server weight success"))
	})
}

// GetServerInfo 从Etcd 获取grpc服务信息
func (r *Register) GetServerInfo() (Server, error) {
	resp, err := r.cli.Get(context.Background(), BuildRegPath(r.srvInfo)) // "/name/version/addr"
	if err != nil {
		return r.srvInfo, err
	}
	info := Server{}
	if resp.Count >= 1 { // 理论上每个grpc服务就一个 key-value，存在多个取匹配的那一个
		if err := json.Unmarshal(resp.Kvs[0].Value, &info); err != nil {
			return info, err
		}
	}
	return info, nil
}
