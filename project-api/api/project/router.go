package project

import (
	"log"
	"project-api/api/middleware"
	"project-api/router"

	"github.com/gin-gonic/gin"
)

// 路由初始化
func init() {
	log.Printf("init project router")
	pu := &RouterProject{}
	router.Register(pu)
}

type RouterProject struct {
}

func (pu *RouterProject) Route(r *gin.Engine) {
	// 初始化Project的Grpc Client=》 ProjectGrpcClinet
	InitProjectGrpcClient()
	//注册验证码函数
	h := NewHandlerProject()
	group := r.Group("project/index")
	group.Use(middleware.TokenVerify())
	group.POST("", h.index)
}
