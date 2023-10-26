package main

import (
	srv "project-common"
	_ "project-user/api" //初始化"user/api/user"路径下的router file
	"project-user/config"
	"project-user/router"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//路由注册
	router.InitRouter(r)

	// Project grpc 连接注册 //初始化rpc调用
	router.InitProjectGrpc()

	//gRPC注册
	gc := router.RegisterGrpc()

	//etcd 注冊
	router.RegisterEtcd()

	//用于grpc 优雅退出
	stop := func() {
		gc.Stop()
	}
	srv.Run(r, config.AppConf.Sc.Name, config.AppConf.Sc.Addr, stop)
}
