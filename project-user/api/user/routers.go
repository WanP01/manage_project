package user

import (
	"log"
	"project-user/router"

	"github.com/gin-gonic/gin"
)

// 路由初始化
func init() {
	log.Printf("init user router")
	ro := &RouterUser{}
	router.Register(ro)
}

// 对 manage_project/user/router/router.go 中Router接口的具体实现
type RouterUser struct {
}

func (ru *RouterUser) Route(r *gin.Engine) {
	hu := NewHandlerUser()
	r.POST("/project/login/getCaptcha", hu.getCaptcha)
}
