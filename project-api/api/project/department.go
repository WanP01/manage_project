package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/api/grpc"
	"project-api/pkg/model/depart"
	common "project-common"
	"project-common/errs"
	"project-grpc/department"
)

type HandlerDepartment struct {
}

func NewHandlerDepartment() *HandlerDepartment {
	return &HandlerDepartment{}
}

func (hd *HandlerDepartment) department(ctx *gin.Context) {
	result := &common.Result{}
	var req *depart.DepartmentReq
	ctx.ShouldBind(&req)
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &department.DepartmentReqMessage{
		Page:                 req.Page,
		PageSize:             req.PageSize,
		ParentDepartmentCode: req.Pcode,
		OrganizationCode:     ctx.GetString("organizationCode"),
	}
	listDepartmentMessage, err := grpc.DepartmentGrpcClient.List(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*depart.Department
	copier.Copy(&list, listDepartmentMessage.List)
	if list == nil {
		list = []*depart.Department{}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"total": listDepartmentMessage.Total,
		"page":  req.Page,
		"list":  list,
	}))
}

func (hd *HandlerDepartment) save(ctx *gin.Context) {
	result := &common.Result{}
	var req *depart.DepartmentReq
	ctx.ShouldBind(&req)
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &department.DepartmentReqMessage{
		Name:                 req.Name,
		DepartmentCode:       req.DepartmentCode,
		ParentDepartmentCode: req.ParentDepartmentCode,
		OrganizationCode:     ctx.GetString("organizationCode"),
	}
	departmentMessage, err := grpc.DepartmentGrpcClient.Save(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var res = &depart.Department{}
	copier.Copy(res, departmentMessage)
	ctx.JSON(http.StatusOK, result.Success(res))
}

func (hd *HandlerDepartment) read(ctx *gin.Context) {
	result := &common.Result{}
	departmentCode := ctx.PostForm("departmentCode")
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &department.DepartmentReqMessage{
		DepartmentCode:   departmentCode,
		OrganizationCode: ctx.GetString("organizationCode"),
	}
	departmentMessage, err := grpc.DepartmentGrpcClient.Read(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var res = &depart.Department{}
	copier.Copy(res, departmentMessage)
	ctx.JSON(http.StatusOK, result.Success(res))
}
