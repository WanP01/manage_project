package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/api/grpc"
	"project-api/pkg/model/auths"
	common "project-common"
	"project-common/errs"
	"project-grpc/account"
)

type HandlerAccount struct {
}

func NewHandlerAccount() *HandlerAccount {
	return &HandlerAccount{}
}

func (a *HandlerAccount) account(ctx *gin.Context) {
	//接收请求参数  一些参数的校验 可以放在api这里
	result := &common.Result{}
	var req *auths.AccountReq
	ctx.ShouldBind(&req)
	memberId := ctx.GetInt64("memberId")
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()

	//调用project模块 查询账户列表
	msg := &account.AccountReqMessage{
		MemberId:         memberId,
		OrganizationCode: ctx.GetString("organizationCode"),
		Page:             int64(req.Page),
		PageSize:         int64(req.PageSize),
		SearchType:       int32(req.SearchType),
		DepartmentCode:   req.DepartmentCode,
	}
	response, err := grpc.AccountGrpcClient.Account(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	//返回数据
	var list []*auths.MemberAccount
	copier.Copy(&list, response.AccountList)
	if list == nil {
		list = []*auths.MemberAccount{}
	}
	var authList []*auths.ProjectAuth
	copier.Copy(&authList, response.AuthList)
	if authList == nil {
		authList = []*auths.ProjectAuth{}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"total":    response.Total,
		"page":     req.Page,
		"list":     list,
		"authList": authList,
	}))
}
