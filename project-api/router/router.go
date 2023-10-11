package router

import (
	"github.com/gin-gonic/gin"
)

// 定义路由接口
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
