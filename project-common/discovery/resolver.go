package discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

const (
	schema = "etcd"
)

// Resolver 整体起到类似DNS 域名解析功能

// Resolver for grpc client
// 客户端连接信息结构体
type Resolver struct {
	schema      string   //模式或主题，前缀名
	EtcdAddrs   []string // etcd 节点地址
	DialTimeout int      // 连接超时时间

	closeCh      chan struct{}      // resolver 关闭端口
	watchCh      clientv3.WatchChan // etcd 的变更监测反馈结果通道
	cli          *clientv3.Client   // etcd 连接客户端
	keyPrifix    string             // key 前缀 "/name/version/" or "/name/"
	srvAddrsList []resolver.Address // 对应 keyPrifix 的 grpc 缓存地址

	cc     resolver.ClientConn // 注册器模块内部注入的连接器（见Build & Register）
	logger *zap.Logger         // 日志
}

// NewResolver create a new resolver.Builder base on etcd
func NewResolver(etcdAddrs []string, logger *zap.Logger) *Resolver {
	return &Resolver{
		schema:      schema,
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// ———————————————————————————————————满足 resolver Bulider 接口（Build && Scheme）———————————————————————————————————————

// Scheme returns the scheme supported by this resolver.
func (r *Resolver) Scheme() string {
	return r.schema
}

// Build creates a new resolver.Resolver for the given target
// 解析 grpc URI语法 // dns:[//authority/]host[:port] =》scheme:opaque[?query][#fragment] or [scheme:][//[userinfo@]host][/]path[?query][#fragment]
func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc
	// 通过grpc 的调用URI 解析出对应参数 与 实际 etcd 中的存储 key 格式一致
	r.keyPrifix = BuildPrefix(Server{Name: target.Endpoint(), Version: strings.Trim(target.URL.Opaque, "/")})

	fmt.Printf("target.URL.Path: %v\n", target.URL.Path)
	fmt.Printf("strings.Trim(target.URL.Opaque, \"/\"): %s\n", strings.Trim(target.URL.Opaque, "/"))
	fmt.Printf("r.keyPrifix: %v\n", r.keyPrifix)

	//启动 register.Start() => 同步etcd key-value && 实时监测并更新
	if _, err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

// ————————————————————————————————————满足 resolver.Resolver interface—————————————————————————————————————————————————

// ResolveNow resolver.Resolver interface
func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

// Close resolver.Resolver interface
func (r *Resolver) Close() {
	r.closeCh <- struct{}{}
}

// ————————————————————————————————————————————————————————————————————————————————————————————————————————————————————

// start
// 注册+同步+监测+更新 register grpc addrs list
func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	// etcd 连接客户端
	r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	//注册注册器，resolver 解析时会按照 Scheme：Bulider 格式查找map
	resolver.Register(r)

	r.closeCh = make(chan struct{})

	// 获取所有r.keyPrifix所对应的 grpc 服务地址（一个resolver 对应一种 grpc服务（以 Name & version 区分））
	if err = r.sync(); err != nil {
		return nil, err
	}

	fmt.Printf("r.srvAddrsList: %v\n", r.srvAddrsList)

	//时刻观测etcd变更，同步更新变更在resolver 的缓存地址上
	go r.watch()

	return r.closeCh, nil
}

// watch update events
// 观测 etcd 的变动
func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)

	// （type WatchChan <-chan WatchResponse）
	r.watchCh = r.cli.Watch(context.Background(), r.keyPrifix, clientv3.WithPrefix())

	for {
		select {
		case <-r.closeCh: // 停止监测etcd，不在更新地址
			return
		case res, ok := <-r.watchCh: //有 key相关的 event 变动会返回 watchresponse （type WatchChan <-chan WatchResponse）
			if ok {
				r.update(res.Events) // 提取 watchresponse 内的 Events 数组 Events []*Event
				fmt.Printf("r.srvAddrsList: %v\n", r.srvAddrsList)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil { // 定时全面更新同步 etcd 的 key-value 信息（会清空旧有 grpc 缓存地址）
				r.logger.Error("sync failed", zap.Error(err))
			}
		}
	}
}

// update
// 根据watch 观测结果更新 resolver 的grpc缓存地址
func (r *Resolver) update(events []*clientv3.Event) {
	for _, ev := range events { //遍历Events=》event
		var info Server
		var err error

		switch ev.Type { // event （put or delete 两种状态）
		case mvccpb.PUT:
			info, err = ParseValue(ev.Kv.Value) // put 包含 最新key-value
			if err != nil {
				continue
			}
			//更新resolver 内的 grpc 服务缓存addr
			addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
			if !Exist(r.srvAddrsList, addr) {
				r.srvAddrsList = append(r.srvAddrsList, addr)
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		case mvccpb.DELETE:
			info, err = SplitPath(string(ev.Kv.Key)) // delete 包含被删除的key-value
			if err != nil {
				continue
			}
			// 清除resolver 内的 grpc 服务缓存addr
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.srvAddrsList, addr); ok {
				r.srvAddrsList = s
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		}
	}
}

// sync 同步获取所有对应grpc的服务地址信息
func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 从etcd 获得以 "/name/verison/" 为前缀的所有 相关 grpc key-value 对
	res, err := r.cli.Get(ctx, r.keyPrifix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	//清空就有信息
	r.srvAddrsList = []resolver.Address{}

	// 将value (key("/name/version/addr" : value ( Server{name,version,addr,weight}))) 还原为 Server 结构体
	for _, v := range res.Kvs {
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}
		//将所有满足前缀的grpc 地址添加到 resolver 的 地址列表中
		addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
		r.srvAddrsList = append(r.srvAddrsList, addr)
	}

	// 将可提供服务的地址列表刷新在 resolver.State 中
	r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
	return nil
}
