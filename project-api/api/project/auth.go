package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/api/grpc"
	"project-api/pkg/model"
	"project-api/pkg/model/auths"
	common "project-common"
	"project-common/errs"
	"project-grpc/auth"
)

type HandlerAuth struct {
}

func (ha *HandlerAuth) authList(ctx *gin.Context) {
	result := &common.Result{}
	organizationCode := ctx.GetString("organizationCode")
	var page = &model.Page{}
	page.Bind(ctx)
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &auth.AuthReqMessage{
		OrganizationCode: organizationCode,
		Page:             page.Page,
		PageSize:         page.PageSize,
	}
	response, err := grpc.AuthGrpcClient.AuthList(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var authList []*auths.ProjectAuth
	copier.Copy(&authList, response.List)
	if authList == nil {
		authList = []*auths.ProjectAuth{}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"total": response.Total,
		"list":  authList,
		"page":  page.Page,
	}))
}

func NewAuth() *HandlerAuth {
	return &HandlerAuth{}
}
