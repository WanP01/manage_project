package user

import (
	"log"
	"project-api/router"

	"github.com/gin-gonic/gin"
)

// 路由初始化
func init() {
	log.Printf("init user router")
	ru := &RouterUser{}
	router.Register(ru)
}

type RouterUser struct {
}

func (ru *RouterUser) Route(r *gin.Engine) {
	// 初始化User的Grpc Client=》 UserGrpcClinet
	InitUserGrpcClient()
	//注册验证码函数
	h := NewHandlerUser()
	r.POST("/project/login/getCaptcha", h.getCaptcha) // 该函数调用User模块Grpc的验证码服务
	r.POST("/project/login/register", h.register)
	r.POST("project/login", h.login)
}
