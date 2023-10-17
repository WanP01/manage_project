package user

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"project-api/config"
	"project-common/discovery"
	"project-common/logs"
	"project-grpc/user/login"
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
