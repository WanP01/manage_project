package user

import (
	"context"
	"fmt"
	"net/http"
	common "project-common"
	"project-common/errs"
	"time"

	login_service_v1 "project-user/pkg/service/login.service.v1"

	"github.com/gin-gonic/gin"
)

type HandlerUser struct {
}

func NewHandlerUser() *HandlerUser {
	return &HandlerUser{}
}

func (hu *HandlerUser) getCaptcha(ctx *gin.Context) {
	result := &common.Result{}
	mobile := ctx.PostForm("mobile")
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//调用User模块的Grpc（验证码服务）
	res, err := UserGrpcClient.GetCaptcha(c, &login_service_v1.CaptchaMessage{Mobile: mobile})
	fmt.Printf("err:%v", err)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	//包装获得的数据反馈
	ctx.JSON(http.StatusOK, result.Success(res.GetCode()))
}
