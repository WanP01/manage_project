package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/api/grpc"
	"project-grpc/project"

	"project-api/pkg/model/pro"
	common "project-common"
	"project-common/errs"
)

func ProjectAuth() func(*gin.Context) {
	return func(ctx *gin.Context) {
		//如果此用户 不是项目的成员 认为你不能查看 不能操作此项目 直接报无权限
		result := &common.Result{}
		//在接口有权限的基础上，做项目权限，不是这个项目的成员，无权限查看项目和操作项目
		//检查是否有projectCode和taskCode这两个参数
		isProjectAuth := false
		projectCode := ctx.PostForm("projectCode")
		if projectCode != "" {
			isProjectAuth = true
		}
		taskCode := ctx.PostForm("taskCode")
		if taskCode != "" {
			isProjectAuth = true
		}
		if isProjectAuth {
			memberId := ctx.GetInt64("memberId")
			//p := project.NewHandlerProject()
			//pr, isMember, isOwner, err := p.FindProjectByMemberId(memberId, projectCode, taskCode)

			//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			//defer cancel()
			c := context.Background()
			msg := &project.ProjectRpcMessage{
				MemberId:    memberId,
				ProjectCode: projectCode,
				TaskCode:    taskCode,
			}
			//查询是否为项目成员
			var pr *pro.Project
			var isMember bool
			var isOwner bool

			projectResponse, err := grpc.ProjectGrpcClient.FindProjectByMemberId(c, msg)

			//数据库错误
			if err != nil {
				code, msg := errs.ParseGrpcError(err)
				ctx.JSON(http.StatusOK, result.Fail(code, msg))
				ctx.Abort()
				return
			}

			// 非项目成员（无操作权限）
			if projectResponse.Project == nil {
				ctx.JSON(http.StatusOK, result.Fail(403, "不是项目成员，无操作权限"))
				ctx.Abort()
				return
			}

			pr = &pro.Project{}
			copier.Copy(pr, projectResponse.Project)
			isOwner = projectResponse.IsOwner
			isMember = projectResponse.IsMember

			if pr.Private == 1 { //私有项目
				if isOwner || isMember { //是所有者或者成员
					ctx.Next()
					return
				} else { //非所有者或成员
					ctx.JSON(http.StatusOK, result.Fail(403, "私有项目，无操作权限"))
					ctx.Abort()
					return
				}
			} else { // 公有项目
				ctx.Next()
				return
			}
		}
	}
}
