package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"os"
	"path"
	"project-api/api/grpc"
	"project-api/pkg/model"
	"project-api/pkg/model/file"
	"project-api/pkg/model/pro"
	"project-api/pkg/model/project_log"
	"project-api/pkg/model/tasks"
	common "project-common"
	"project-common/errs"
	"project-common/fs"
	"project-common/tms"
	"project-grpc/task"
	"project-user/config"
	"time"
)

type HandlerTask struct {
}

func NewHandleTask() *HandlerTask {
	return &HandlerTask{}
}

func (ht *HandlerTask) taskStages(ctx *gin.Context) {
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

func (ht *HandlerTask) memberProjectList(ctx *gin.Context) {
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
	var list []*pro.MemberProjectResp
	err = copier.Copy(&list, resp.List)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if list == nil { // 赋予默认值
		list = []*pro.MemberProjectResp{}
	}

	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"total": resp.Total,
		"page":  page.Page,
		"list":  list,
	}))
}

func (ht *HandlerTask) taskList(ctx *gin.Context) {
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

func (ht *HandlerTask) taskSave(ctx *gin.Context) {
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

func (ht *HandlerTask) taskSort(ctx *gin.Context) {
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

func (ht *HandlerTask) myTaskList(ctx *gin.Context) {
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

func (ht *HandlerTask) taskRead(ctx *gin.Context) {
	result := &common.Result{}
	taskCode := ctx.PostForm("taskCode")
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: ctx.GetInt64("memberId"),
	}
	taskMessage, err := grpc.TaskGrpcClient.TaskRead(c, msg)
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
	ctx.JSON(200, result.Success(td))
}

func (ht *HandlerTask) listTaskMember(ctx *gin.Context) {
	result := &common.Result{}
	taskCode := ctx.PostForm("taskCode")
	page := &model.Page{}
	page.Bind(ctx)
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: ctx.GetInt64("memberId"),
		Page:     page.Page,
		PageSize: page.PageSize,
	}
	taskMemberResponse, err := grpc.TaskGrpcClient.ListTaskMember(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var tms []*tasks.TaskMember
	copier.Copy(&tms, taskMemberResponse.List)
	if tms == nil {
		tms = []*tasks.TaskMember{}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"list":  tms,
		"total": taskMemberResponse.Total,
		"page":  page.Page,
	}))
}

func (ht *HandlerTask) taskLog(ctx *gin.Context) {
	result := &common.Result{}
	var req *project_log.TaskLogReq
	ctx.ShouldBind(&req)
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &task.TaskReqMessage{
		TaskCode: req.TaskCode,
		MemberId: ctx.GetInt64("memberId"),
		Page:     int64(req.Page),
		PageSize: int64(req.PageSize),
		All:      int32(req.All),
		Comment:  int32(req.Comment),
	}
	taskLogResponse, err := grpc.TaskGrpcClient.TaskLog(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var tms []*project_log.ProjectLogDisplay
	copier.Copy(&tms, taskLogResponse.List)
	if tms == nil {
		tms = []*project_log.ProjectLogDisplay{}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"list":  tms,
		"total": taskLogResponse.Total,
		"page":  req.Page,
	}))
}

func (ht *HandlerTask) taskWorkTimeList(ctx *gin.Context) {
	taskCode := ctx.PostForm("taskCode")
	result := &common.Result{}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &task.TaskReqMessage{
		TaskCode: taskCode,
		MemberId: ctx.GetInt64("memberId"),
	}
	taskWorkTimeResponse, err := grpc.TaskGrpcClient.TaskWorkTimeList(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var tms []*tasks.TaskWorkTime
	copier.Copy(&tms, taskWorkTimeResponse.List)
	if tms == nil {
		tms = []*tasks.TaskWorkTime{}
	}
	ctx.JSON(http.StatusOK, result.Success(tms))
}

func (ht *HandlerTask) saveTaskWorkTime(ctx *gin.Context) {
	result := &common.Result{}
	var req *tasks.SaveTaskWorkTimeReq
	ctx.ShouldBind(&req)
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	msg := &task.TaskReqMessage{
		TaskCode:  req.TaskCode,
		MemberId:  ctx.GetInt64("memberId"),
		Content:   req.Content,
		Num:       int32(req.Num),
		BeginTime: tms.ParseTime(req.BeginTime),
	}
	_, err := grpc.TaskGrpcClient.SaveTaskWorkTime(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	ctx.JSON(http.StatusOK, result.Success([]int{}))
}

func (ht *HandlerTask) uploadFiles(ctx *gin.Context) {
	result := &common.Result{}
	req := file.UploadFileReq{}
	ctx.ShouldBind(&req)
	//处理文件
	multipartForm, _ := ctx.MultipartForm()
	file := multipartForm.File
	//假设只上传一个文件
	uploadFile := file["file"][0]
	//第一种 没有达成分片的条件
	key := ""
	if req.TotalChunks == 1 {
		//不分片，不需要处理拼接，直接存储就行
		// path upload/项目Id/任务Id/时间戳/Filename
		path := "upload/" + req.ProjectCode + "/" + req.TaskCode + "/" + tms.FormatYMD(time.Now())
		if !fs.IsExist(path) { //判断路径是否存在
			os.MkdirAll(path, os.ModePerm) // 不存在就新建路径
		}
		dst := path + "/" + req.Filename
		key = dst //upload/项目Id/任务Id/时间戳/Filename
		err := ctx.SaveUploadedFile(uploadFile, dst)
		if err != nil {
			ctx.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
	}
	if req.TotalChunks > 1 {
		//分片上传 无非就是先把每次的存储起来 追加就可以了
		path := "upload/" + req.ProjectCode + "/" + req.TaskCode + "/" + tms.FormatYMD(time.Now())
		if !fs.IsExist(path) {
			os.MkdirAll(path, os.ModePerm)
		}
		fileName := path + "/" + req.Identifier                                                // upload/项目Id/任务Id/时间戳/identifier
		openFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm) //创建并打开目标存储文件
		if err != nil {
			ctx.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		// 打开上传文件的内容
		open, err := uploadFile.Open()
		if err != nil {
			ctx.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		defer open.Close()
		//读取上传文件
		buf := make([]byte, req.CurrentChunkSize)
		open.Read(buf)
		//写入目标文件并关闭
		openFile.Write(buf)
		openFile.Close()
		//如果是最后一个，那么将名字改为filename
		key = fileName //upload/项目Id/任务Id/时间戳/identifier

		// upload/项目Id/任务Id/时间戳/identifier =》 upload/项目Id/任务Id/时间戳/Filename
		if req.TotalChunks == req.ChunkNumber {
			//最后一个分片了
			newPath := path + "/" + req.Filename
			key = newPath
			os.Rename(fileName, newPath)
		}
	}
	//调用服务 存入file表 ms_file && ms_source_link
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()

	fileUrl := model.HttpProtocol + config.AppConf.Sc.Addr + "/" + key

	//最后一次 相关文件信息 才保存在mysql中
	if req.TotalChunks == req.ChunkNumber {
		msg := &task.TaskFileReqMessage{
			TaskCode:         req.TaskCode,
			ProjectCode:      req.ProjectCode,
			OrganizationCode: ctx.GetString("organizationCode"),
			PathName:         key,
			FileName:         req.Filename,
			Size:             int64(req.TotalSize),
			Extension:        path.Ext(key),
			FileUrl:          fileUrl,
			FileType:         file["file"][0].Header.Get("Content-Type"),
			MemberId:         ctx.GetInt64("memberId"),
		}

		_, err := grpc.TaskGrpcClient.SaveTaskFile(ctx, msg)
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			ctx.JSON(http.StatusOK, result.Fail(code, msg))
		}
	}

	// 但每个分片都需要返回响应
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"file":        key, // 文件标记 upload/项目Id/任务Id/时间戳/filename
		"hash":        "",
		"key":         key,
		"url":         fileUrl, // 文件URL http://localhost/upload/项目Id/任务Id/时间戳/filename
		"projectName": req.ProjectName,
	}))
	return
}

func (ht *HandlerTask) taskSources(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sources, err := grpc.TaskGrpcClient.TaskSources(ctx, &task.TaskReqMessage{TaskCode: taskCode})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var slList []*file.SourceLink
	copier.Copy(&slList, sources.List)
	if slList == nil {
		slList = []*file.SourceLink{}
	}
	c.JSON(http.StatusOK, result.Success(slList))
}

func (ht *HandlerTask) createComment(c *gin.Context) {
	result := &common.Result{}
	req := model.CommentReq{}
	c.ShouldBind(&req)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &task.TaskReqMessage{
		TaskCode:       req.TaskCode,
		CommentContent: req.Comment,
		Mentions:       req.Mentions,
		MemberId:       c.GetInt64("memberId"),
	}
	_, err := grpc.TaskGrpcClient.CreateComment(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success(true))
}
