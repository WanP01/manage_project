package router

import (
	"log"
	"net"
	"project-common/discovery"
	"project-common/logs"
	"project-grpc/user/login"
	"project-user/config"
	LoginServiceV1 "project-user/pkg/service/login.service.v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

// Router 定义路由接口
type Router interface {
	Route(r *gin.Engine)
}

// //方案1:注册函数
// // 用于批量注册路由Router实现
// type RegisterRouter struct {
// }

// func NewRegisterRouter() *RegisterRouter {
// 	return &RegisterRouter{}
// }

// // 路由批量注册器：实现Router接口，调用单个路由实现
// func (rr *RegisterRouter) Route(ro Router, r *gin.Engine) {
// 	ro.Route(r)
// }

// 方案2:路由Router列表
// 在manage_project/user/api/userrouter/routers.go init 函数初始化
var routers []Router

func Register(arg ...Router) {
	routers = append(routers, arg...)
}

func InitRouter(r *gin.Engine) {
	//方案1
	// 一列列注册路由
	// rg := NewRegisterRouter()
	// rg.Route(&user.RouterUser{}, r)

	// 方案2
	for _, ro := range routers {
		ro.Route(r)
	}

}

// 普通grpc注册的信息
type gRPCConfig struct {
	Addr         string
	RegisterFunc func(*grpc.Server)
}

// RegisterGrpc 注册并启动grpc服务
func RegisterGrpc() *grpc.Server {
	c := &gRPCConfig{
		Addr: config.AppConf.Gc.Addr,
		RegisterFunc: func(s *grpc.Server) {
			login.RegisterLoginServiceServer(s, LoginServiceV1.New())
		}}
	s := grpc.NewServer()
	c.RegisterFunc(s)
	lis, err := net.Listen("tcp", c.Addr)
	if err != nil {
		log.Println("cannot listen")
	}
	go func() {
		log.Printf("grpc server started as: %s \n", c.Addr)
		err := s.Serve(lis)
		if err != nil {
			log.Println("server started error", err)
			return
		}
	}()
	return s
}

// RegisterEtcd 注册 Etcd 服务(Etcd服务单独启动)
func RegisterEtcd() {
	//注册grpc URL解析器
	etcdRegister := discovery.NewResolver(config.AppConf.Ec.Addrs, logs.LG)
	resolver.Register(etcdRegister)
	// 构建 grpc 服务Server信息
	srvInfo := discovery.Server{
		Name: config.AppConf.Gc.Name,
		Addr: config.AppConf.Gc.Addr,
	}
	//在Etcd 上注册服务信息
	r := discovery.NewRegister(config.AppConf.Ec.Addrs, logs.LG)
	_, err := r.Register(srvInfo, 5)
	if err != nil {
		log.Fatalln(err)
	}
}
