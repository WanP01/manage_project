package project

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/api/grpc"
	"project-api/pkg/model"
	"project-api/pkg/model/auths"
	"project-api/pkg/model/pro"
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

func (ha *HandlerAuth) apply(ctx *gin.Context) {
	result := &common.Result{}
	var req *auths.ProjectAuthReq
	ctx.ShouldBind(&req)
	var nodes []string
	if req.Nodes != "" {
		json.Unmarshal([]byte(req.Nodes), &nodes)
	}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &auth.AuthReqMessage{
		Action: req.Action,
		AuthId: req.Id,
		Nodes:  nodes,
	}
	applyResponse, err := grpc.AuthGrpcClient.Apply(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*pro.ProjectNodeAuthTree
	copier.Copy(&list, applyResponse.List)
	var checkedList []string
	copier.Copy(&checkedList, applyResponse.CheckedList)
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"list":        list,
		"checkedList": checkedList,
	}))
}

//func (ha *HandlerAuth) GetAuthNodes(ctx *gin.Context) ([]string, error) {
//	memberId := ctx.GetInt64("memberId")
//	msg := &auth.AuthReqMessage{
//		MemberId: memberId,
//	}
//	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//	//defer cancel()
//	c := context.Background()
//	response, err := grpc.AuthGrpcClient.AuthNodesByMemberId(c, msg)
//	if err != nil {
//		code, msg := errs.ParseGrpcError(err)
//		return nil, errs.NewError(errs.ErrorCode(code), msg)
//	}
//	return response.List, err
//}

func NewAuth() *HandlerAuth {
	return &HandlerAuth{}
}
