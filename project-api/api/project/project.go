package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	common "project-common"
	"project-common/errs"
	"project-grpc/project"
)

type HandlerProject struct {
}

func NewHandlerProject() *HandlerProject {
	return &HandlerProject{}
}

func (hp *HandlerProject) index(ctx *gin.Context) {

	// 1.接收参数 参数模型
	result := &common.Result{}

	// 2.调用user grpc 完成登录
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background() // 调试用
	msg := &project.IndexMessage{}
	Resp, err := ProjectGrpcClient.Index(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	// 回复响应的用户数据
	ctx.JSON(http.StatusOK, result.Success(Resp.Menus))
}
