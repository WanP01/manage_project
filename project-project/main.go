package main

import (
	srv "project-common"
	"project-project/config"
	"project-project/router"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//路由注册
	router.InitRouter(r)

	//gRPC注册
	gc := router.RegisterGrpc()

	//etcd 注冊
	router.RegisterEtcd()

	// User grpc 连接注册 //初始化rpc调用
	router.InitUserGrpc()

	//用于grpc 优雅退出
	stop := func() {
		gc.Stop()
	}
	srv.Run(r, config.AppConf.Sc.Name, config.AppConf.Sc.Addr, stop)
}
