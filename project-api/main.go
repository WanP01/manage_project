package main

import (
	"net/http"
	_ "project-api/api" //初始化"user/api/user & /project"路径下的router file
	"project-api/api/middleware"
	"project-api/config"
	"project-api/router"
	srv "project-common"

	"github.com/gin-gonic/gin"
)

// 仅负责api 启动和路由
func main() {
	r := gin.Default()

	r.Use(middleware.RequestLog())

	//静态文件映射
	r.StaticFS("/upload", http.Dir("upload"))
	//路由注册
	router.InitRouter(r)

	srv.Run(r, config.AppConf.Sc.Name, config.AppConf.Sc.Addr, nil)
}
