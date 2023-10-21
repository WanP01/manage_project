package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/api/grpc"
	"project-api/pkg/model"
	"project-api/pkg/model/project"
	"project-api/pkg/model/tasks"
	common "project-common"
	"project-common/errs"
	"project-grpc/task"
)

type HandlerTask struct {
}

func NewHandleTask() *HandlerTask {
	return &HandlerTask{}
}

func (t HandlerTask) taskStages(ctx *gin.Context) {
	result := &common.Result{}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	//1.获取参数 校验参数的合法性
	projectCode := ctx.PostForm("projectCode")
	page := &model.Page{}
	page.Bind(ctx)

	//2.调用grpc服务
	msg := &task.TaskReqMessage{
		MemberId:    ctx.GetInt64("memberId"),
		ProjectCode: projectCode,
		Page:        page.Page,
		PageSize:    page.PageSize,
	}
	stages, err := grpc.TaskGrpcClient.TaskStages(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	//if stages.List == nil { // 赋予默认值
	//	stages.List = []*task.TaskStagesMessage{}
	//}
	var list []*tasks.TaskStagesResp
	err = copier.Copy(&list, stages.List)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if list == nil { // 赋予默认值
		list = []*tasks.TaskStagesResp{}
	}
	for _, v := range list {
		v.TasksLoading = true  //任务加载状态
		v.FixedCreator = false //添加任务按钮定位
		v.ShowTaskCard = false //是否显示创建卡片
		v.Tasks = []int{}
		v.DoneTasks = []int{}
		v.UnDoneTasks = []int{}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"total": stages.Total,
		"page":  page.Page,
		"list":  list,
	}))
}

func (t HandlerTask) memberProjectList(ctx *gin.Context) {
	result := &common.Result{}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	//1.获取参数 校验参数的合法性
	projectCode := ctx.PostForm("projectCode")
	page := &model.Page{}
	page.Bind(ctx)

	//2.调用grpc服务
	msg := &task.TaskReqMessage{
		MemberId:    ctx.GetInt64("memberId"),
		ProjectCode: projectCode,
		Page:        page.Page,
		PageSize:    page.PageSize,
	}
	resp, err := grpc.TaskGrpcClient.MemberProjectList(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	//if resp.List == nil { // 赋予默认值
	//	resp.List = []*task.MemberProjectMessage{}
	//}
	// 3. 拼整数据 grpc => http
	var list []*project.MemberProjectResp
	err = copier.Copy(&list, resp.List)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if list == nil { // 赋予默认值
		list = []*project.MemberProjectResp{}
	}

	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"total": resp.Total,
		"page":  page.Page,
		"list":  list,
	}))
}

func (t HandlerTask) taskList(ctx *gin.Context) {
	result := &common.Result{}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	//1.获取参数 校验参数的合法性
	stageCode := ctx.PostForm("stageCode")
	//2.调用grpc服务
	msg := &task.TaskReqMessage{
		MemberId:  ctx.GetInt64("memberId"),
		StageCode: stageCode,
	}
	resp, err := grpc.TaskGrpcClient.TaskList(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	// 3. 拼整数据 grpc => http
	var list []*tasks.TaskDisplay
	err = copier.Copy(&list, resp.List)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if list == nil { // 赋予默认值
		list = []*tasks.TaskDisplay{}
	}
	for _, v := range list { // 赋予默认值
		if v.Tags == nil {
			v.Tags = []int{}
		}
		if v.ChildCount == nil {
			v.ChildCount = []int{}
		}
	}

	ctx.JSON(http.StatusOK, result.Success(list))
}

func (t *HandlerTask) taskSave(ctx *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSaveReq
	ctx.ShouldBind(&req)
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &task.TaskReqMessage{
		ProjectCode: req.ProjectCode,
		Name:        req.Name,
		StageCode:   req.StageCode,
		AssignTo:    req.AssignTo,
		MemberId:    ctx.GetInt64("memberId"),
	}
	taskMessage, err := grpc.TaskGrpcClient.TaskSave(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	td := &tasks.TaskDisplay{}
	copier.Copy(td, taskMessage)
	if td != nil {
		if td.Tags == nil {
			td.Tags = []int{}
		}
		if td.ChildCount == nil {
			td.ChildCount = []int{}
		}
	}
	ctx.JSON(http.StatusOK, result.Success(td))
}

func (t *HandlerTask) taskSort(ctx *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSortReq
	ctx.ShouldBind(&req)
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &task.TaskReqMessage{
		PreTaskCode:  req.PreTaskCode,
		NextTaskCode: req.NextTaskCode,
		ToStageCode:  req.ToStageCode,
		MemberId:     ctx.GetInt64("memberId"),
	}
	_, err := grpc.TaskGrpcClient.TaskSort(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}

	ctx.JSON(http.StatusOK, result.Success([]int{}))
}

func (t *HandlerTask) myTaskList(ctx *gin.Context) {
	result := &common.Result{}
	var req *tasks.MyTaskReq
	ctx.ShouldBind(&req)
	memberId := ctx.GetInt64("memberId")
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &task.TaskReqMessage{
		MemberId: memberId,
		TaskType: int32(req.TaskType),
		Type:     int32(req.Type),
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	myTaskListResponse, err := grpc.TaskGrpcClient.MyTaskList(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var myTaskList []*tasks.MyTaskDisplay
	copier.Copy(&myTaskList, myTaskListResponse.List)
	if myTaskList == nil {
		myTaskList = []*tasks.MyTaskDisplay{}
	}
	for _, v := range myTaskList {
		v.ProjectInfo = tasks.ProjectInfo{
			Name: v.ProjectName,
			Code: v.ProjectCode,
		}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"list":  myTaskList,
		"total": myTaskListResponse.Total,
	}))
}
