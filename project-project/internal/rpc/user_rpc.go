package rpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"project-common/discovery"
	"project-common/logs"
	"project-grpc/user/login"
	"project-project/config"
)

// UserGrpcClient 全局变量（方便复用）
var UserGrpcClient login.LoginServiceClient

func InitUserGrpcClient() {
	// 注册grpc resolver 解析器解析 URL
	etcdRegister := discovery.NewResolver(config.AppConf.Ec.Addrs, logs.LG)
	resolver.Register(etcdRegister)
	// 连接GRPC端口
	conn, err := grpc.Dial("etcd:///user", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect:%v", err)
	}
	UserGrpcClient = login.NewLoginServiceClient(conn)
}
