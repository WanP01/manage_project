package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"project-api/api/grpc"
	common "project-common"
	"project-common/errs"
	"project-grpc/auth"
	"strings"
)

var ignores = []string{
	"project/login/register",
	"project/login",
	"project/login/getCaptcha",
	"project/organization",
	"project/auth/apply"}

func NodeAuthCheck() func(*gin.Context) {
	return func(ctx *gin.Context) {
		result := &common.Result{}
		uri := ctx.Request.RequestURI
		// 先确认忽略列表
		for _, v := range ignores {
			if strings.Contains(uri, v) {
				ctx.Next()
				return
			}
		}

		//判断此uri是否在用户的授权列表中
		//a := NewAuth()
		//nodes, err := a.GetAuthNodes(ctx)
		memberId := ctx.GetInt64("memberId")
		msg := &auth.AuthReqMessage{
			MemberId: memberId,
		}
		//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		//defer cancel()
		c := context.Background()
		nodes, err := grpc.AuthGrpcClient.AuthNodesByMemberId(c, msg)

		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			ctx.JSON(http.StatusOK, result.Fail(code, msg))
			ctx.Abort()
			return
		}

		for _, v := range nodes.List {
			if strings.Contains(uri, v) {
				ctx.Next()
				return
			}
		}

		ctx.JSON(http.StatusOK, result.Fail(403, "无权限操作"))
		ctx.Abort()
		return

	}
}
