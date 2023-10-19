package project

import (
	"log"
	"project-api/api/grpc"
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
	grpc.InitProjectGrpcClient()
	//注册验证码函数
	hp := NewHandlerProject()
	group := r.Group("/project")
	group.Use(middleware.TokenVerify())
	group.POST("/index", hp.index)
	group.POST("/project/selfList", hp.myProjectList)
	group.POST("/project", hp.myProjectList)
	group.POST("/project_template", hp.projectTemplate)
	group.POST("/project/save", hp.projectSave)
	group.POST("/project/read", hp.projectRead)
	group.POST("/project/recycle", hp.projectRecycle)
	group.POST("/project/recovery", hp.projectRecovery)
	group.POST("/project_collect/collect", hp.projectCollect)
	group.POST("/project/edit", hp.projectEdit)
}
