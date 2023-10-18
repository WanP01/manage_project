package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/api/grpc"
	"project-api/pkg/model"
	"project-api/pkg/model/menu"
	pro "project-api/pkg/model/project"
	common "project-common"
	"project-common/errs"
	"project-grpc/project"
	"strconv"
	"time"
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
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//c := context.Background() // 调试用
	msg := &project.IndexMessage{}
	Resp, err := grpc.ProjectGrpcClient.Index(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	// 回复响应的用户数据，变更格式
	var ms []*menu.Menu
	err = copier.Copy(&ms, Resp.Menus)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(ms))
}

func (hp *HandlerProject) myProjectList(ctx *gin.Context) {
	result := &common.Result{}
	//1. 获取参数
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//c := context.Background() // 调试用

	memId := ctx.GetInt64("memberId")
	memberName := ctx.GetString("memberName")
	page := &model.Page{}
	page.Bind(ctx)
	selectBy := ctx.PostForm("selectBy")
	// 2.调用user grpc 完成登录
	msg := &project.ProjectRpcMessage{
		MemberId:   memId,
		Page:       page.Page,
		PageSize:   page.PageSize,
		MemberName: memberName,
		SelectBy:   selectBy,
	}
	myProjectResponse, err := grpc.ProjectGrpcClient.FindProjectByMemId(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var pam []*pro.ProjectAndMember
	err = copier.Copy(&pam, myProjectResponse.Pm)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if pam == nil { // 如果不存在值，应当赋予默认值 //null nil -> []
		pam = []*pro.ProjectAndMember{}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"list":  pam,
		"total": myProjectResponse.Total,
	}))
}

func (hp *HandlerProject) projectTemplate(ctx *gin.Context) {
	result := &common.Result{}
	//1. 获取参数
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	memberId := ctx.GetInt64("memberId")
	memberName := ctx.GetString("memberName")
	page := &model.Page{}
	page.Bind(ctx)
	viewTypeStr := ctx.PostForm("viewType")
	viewType, _ := strconv.ParseInt(viewTypeStr, 10, 64)
	msg := &project.ProjectRpcMessage{
		MemberId:         memberId,
		MemberName:       memberName,
		ViewType:         int32(viewType),
		Page:             page.Page,
		PageSize:         page.PageSize,
		OrganizationCode: ctx.GetString("organizationCode")}
	templateResponse, err := grpc.ProjectGrpcClient.FindProjectTemplate(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}

	var pms []*pro.ProjectTemplate
	err = copier.Copy(&pms, templateResponse.Ptm)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if pms == nil {
		pms = []*pro.ProjectTemplate{}
	}
	for _, v := range pms {
		if v.TaskStages == nil {
			v.TaskStages = []*pro.TaskStagesOnlyName{}
		}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"list":  pms, //null nil -> []
		"total": templateResponse.Total,
	}))
}

func (hp *HandlerProject) projectSave(ctx *gin.Context) {
	result := &common.Result{}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	memberId := ctx.GetInt64("memberId")
	memberName := ctx.GetString("memberName")
	organizationCode := ctx.GetString("organizationCode")
	//projectName := ctx.GetString("name")
	//templateCode := ctx.GetString("templateCode")
	//description := ctx.GetString("description")
	//projectId := ctx.GetInt64("id")
	saveReq := pro.SaveProjectRequest{}
	ctx.ShouldBind(&saveReq)
	msg := &project.ProjectRpcMessage{
		MemberId:         memberId,
		MemberName:       memberName,
		OrganizationCode: organizationCode,
		Name:             saveReq.Name,
		TemplateCode:     saveReq.TemplateCode,
		Description:      saveReq.Description,
		Id:               int64(saveReq.Id),
	}
	projectResp, err := grpc.ProjectGrpcClient.SaveProject(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var sp []*pro.SaveProject
	err = copier.Copy(&sp, projectResp)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(sp))
}
