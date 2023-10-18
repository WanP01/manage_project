package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"project-api/api/grpc"
	common "project-common"
	"project-common/errs"
	"project-grpc/user/login"
)

func TokenVerify() func(*gin.Context) {
	return func(ctx *gin.Context) {
		result := &common.Result{}
		//1.从header中获取token
		token := ctx.GetHeader("Authorization")
		//2.调用user服务进行token认证
		//c, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
		//defer cancelFunc()
		//c := context.Background()
		response, err := grpc.UserGrpcClient.TokenVerify(ctx, &login.LoginMessage{Token: token})
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			ctx.JSON(http.StatusOK, result.Fail(code, msg))
			ctx.Abort()
			return
		}
		//3.处理结果，认证通过 将信息放入gin的上下文 失败返回未登录
		ctx.Set("memberId", response.Member.Id)
		ctx.Set("memberName", response.Member.Name)
		ctx.Set("organizationCode", response.Member.OrganizationCode)
		ctx.Next()
	}
}
