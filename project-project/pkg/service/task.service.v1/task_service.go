package task_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"math"
	"project-common/encrypts"
	"project-common/errs"
	"project-common/tms"
	"project-grpc/task"
	"project-grpc/user/login"
	"project-project/config"
	"project-project/domain"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
	"project-project/pkg/model"
	"time"
)

type TaskService struct {
	task.UnimplementedTaskServiceServer
	cache                  repo.Cache
	transaction            tran.Transaction
	menuRepo               repo.MenuRepo
	projectRepo            repo.ProjectRepo
	projectTemplateRepo    repo.ProjectTemplateRepo
	taskStagesTemplateRepo repo.TaskStagesTemplateRepo
	taskStagesRepo         repo.TaskStagesRepo
	taskRepo               repo.TaskRepo
	projectLogRepo         repo.ProjectLogRepo
	//taskWorkTimeRepo       repo.TaskWorkTimeRepo
	fileRepo           repo.FileRepo
	sourceLinkRepo     repo.SourceLinkRepo
	taskWorkTimeDomain *domain.TaskWorkTimeDomain //依赖注入
	userRpcDomain      *domain.UserRpcDomain      //依赖注入
}

func New() *TaskService {
	return &TaskService{
		cache:                  dao.Rc,
		transaction:            dao.NewTransactionDao(),
		menuRepo:               dao.NewMenuDao(),
		projectRepo:            dao.NewProjectDao(),
		projectTemplateRepo:    dao.NewProjectTemplateDao(),
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(),
		taskStagesRepo:         dao.NewTaskStagesDao(),
		taskRepo:               dao.NewTaskDao(),
		projectLogRepo:         dao.NewProjectLogDao(),
		//taskWorkTimeRepo:       dao.NewTaskWorkTimeDao(),
		fileRepo:           dao.NewFileDao(),
		sourceLinkRepo:     dao.NewSourceLinkDao(),
		userRpcDomain:      domain.NewUserRpcDomain(),
		taskWorkTimeDomain: domain.NewTaskWorkTimeDomain(),
	}
}

func (ts *TaskService) TaskStages(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskStagesResponse, error) {
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	page := msg.Page
	size := msg.PageSize
	stages, total, err := ts.taskStagesRepo.FindTaskByProjectId(ctx, projectCode, page, size)
	if err != nil {
		zap.L().Error("task taskStages taskStagesRepo.FindTaskByProjectId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	var tsMessage []*task.TaskStagesMessage
	copier.Copy(&tsMessage, stages)
	if tsMessage == nil {
		return &task.TaskStagesResponse{List: tsMessage, Total: 0}, nil
	}
	stageMap := data.ToTaskStagesMap(stages)
	for _, v := range tsMessage {
		Id := int(v.Id)
		v.Code = encrypts.EncryptNoErr(int64(v.Id))
		v.ProjectCode = encrypts.EncryptNoErr(stageMap[Id].ProjectCode)
		v.CreateTime = tms.FormatByMill(stageMap[Id].CreateTime)
	}

	return &task.TaskStagesResponse{List: tsMessage, Total: total}, nil
}

func (ts *TaskService) MemberProjectList(ctx context.Context, msg *task.TaskReqMessage) (*task.MemberProjectResponse, error) {
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	page := msg.Page
	size := msg.PageSize
	// 1. 获取项目对应成员ID =》 project_member
	projectMembers, total, err := ts.projectRepo.FindProjectMemberByPId(ctx, projectCode, page, size)
	if err != nil {
		zap.L().Error("task MemberProjectList projectRepo.FindProjectMemberByPId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//2.拿上用户id列表 去请求用户信息（需要先构建 UserRequestMessage）
	if projectMembers == nil || len(projectMembers) <= 0 { // 无成员直接返回
		return &task.MemberProjectResponse{List: nil, Total: 0}, nil
	}

	var mIds []int64
	pmMap := make(map[int64]*data.ProjectMember)
	for _, v := range projectMembers {
		mIds = append(mIds, v.MemberCode)
		pmMap[v.MemberCode] = v
	}

	// 调用User Grpc找到相关member信息
	//userMsg := &login.UserMessage{
	//	MIds: mIds,
	//}
	//memInfoList, err := rpc.UserGrpcClient.FindMemInfoByIds(ctx, userMsg)
	memInfoList, _, err := ts.userRpcDomain.MemberList(ctx, mIds)
	if err != nil {
		zap.L().Error("project MemberProjectList ts.userRpcDomain.MemberList error", zap.Error(err))
		return nil, err
	}

	var list []*task.MemberProjectMessage
	for _, v := range memInfoList.List {
		mpm := &task.MemberProjectMessage{
			Name:       v.Name,
			Avatar:     v.Avatar,
			MemberCode: v.Id,
			Code:       v.Code,
			Email:      v.Email,
		}
		OwnerCode := encrypts.EncryptNoErr(pmMap[v.Id].IsOwner)
		if v.Code == OwnerCode {
			mpm.IsOwner = model.Owner
		}
		list = append(list, mpm)
	}
	return &task.MemberProjectResponse{List: list, Total: total}, nil
}

func (ts *TaskService) TaskList(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskListResponse, error) {
	stageCode := encrypts.DecryptNoErr(msg.StageCode)
	taskList, err := ts.taskRepo.FindTaskByStageCode(ctx, int(stageCode))
	if err != nil {
		zap.L().Error("project task TaskList FindTaskByStageCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var taskDisplayList []*data.TaskDisplay
	var mIds []int64 // 每一项task的指派人Id list
	for _, v := range taskList {
		td := v.ToTaskDisplay()
		if v.Private == 1 { // 代表隐私模式
			tm, err := ts.taskRepo.FindTaskMemberByTaskId(ctx, v.Id, msg.MemberId)
			if err != nil {
				zap.L().Error("project task TaskList FindTaskMemberByTaskId error", zap.Error(err))
				return nil, errs.GrpcError(model.DBError)
			}
			if tm == nil {
				td.CanRead = model.NoCanRead
			} else {
				td.CanRead = model.CanRead
			}
		} else { // 非隐私模式（公开）
			td.CanRead = model.CanRead
		}
		taskDisplayList = append(taskDisplayList, td)
		mIds = append(mIds, v.AssignTo)
	}
	// 赋值Executor 调用user grpc 模块查找 member信息
	if mIds == nil || len(mIds) <= 0 {
		return &task.TaskListResponse{List: nil}, nil
	}
	// 注意 mIds == nil 即 in （null）的情况
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	//c := context.Background()
	//memList, err := rpc.UserGrpcClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIds})
	//if err != nil {
	//	zap.L().Error("project task TaskList LoginServiceClient.FindMemInfoByIds error", zap.Error(err))
	//	return nil, err
	//}
	//memMap := make(map[int64]*login.MemberMessage)
	//for _, v := range memList.List {
	//	memMap[v.Id] = v
	//}
	_, memMap, err := ts.userRpcDomain.MemberList(ctx, mIds)
	if err != nil {
		zap.L().Error("project task TaskList ts.userRpcDomain.MemberList error", zap.Error(err))
		return nil, err
	}

	for _, v := range taskDisplayList {
		memMsg := memMap[encrypts.DecryptNoErr(v.AssignTo)]
		v.Executor = data.Executor{
			Name:   memMsg.Name,
			Avatar: memMsg.Avatar,
		}
	}

	var taskMessageList []*task.TaskMessage
	copier.Copy(&taskMessageList, taskDisplayList)
	return &task.TaskListResponse{List: taskMessageList}, nil
}

func (ts *TaskService) TaskSave(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskMessage, error) {
	//先检查业务
	//确认task 名不能为空
	if msg.Name == "" {
		return nil, errs.GrpcError(model.TaskNameNotNull)
	}
	//确认 任务阶段code 存在
	stageCode := encrypts.DecryptNoErr(msg.StageCode)
	taskStages, err := ts.taskStagesRepo.FindById(ctx, int(stageCode))
	if err != nil {
		zap.L().Error("project task SaveTask taskStagesRepo.FindById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskStages == nil {
		return nil, errs.GrpcError(model.TaskStagesNotNull)
	}
	// 确认对应项目code存在
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	project, err := ts.projectRepo.FindProjectById(ctx, projectCode)
	if err != nil {
		zap.L().Error("project task SaveTask projectRepo.FindProjectById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if project == nil || project.Deleted == model.Deleted {
		return nil, errs.GrpcError(model.ProjectAlreadyDeleted)
	}

	// 确认项目当前最大task num
	maxIdNum, err := ts.taskRepo.FindTaskMaxIdNum(ctx, projectCode)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskMaxIdNum error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if maxIdNum == nil { //预防空指针
		a := 0
		maxIdNum = &a
	}
	//确认项目当前最大 task 排序
	maxSort, err := ts.taskRepo.FindTaskSort(ctx, projectCode, stageCode)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskSort error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if maxSort == nil { // 预防空指针
		a := 0
		maxSort = &a
	}
	//确认指派人
	assignTo := encrypts.DecryptNoErr(msg.AssignTo)

	ta := &data.Task{
		Name:        msg.Name,
		CreateTime:  time.Now().UnixMilli(),
		CreateBy:    msg.MemberId,
		AssignTo:    assignTo,
		ProjectCode: projectCode,
		StageCode:   int(stageCode),
		IdNum:       *maxIdNum + 1,
		Private:     project.OpenTaskPrivate,
		Sort:        *maxSort + 65536,
		BeginTime:   time.Now().UnixMilli(),
		EndTime:     time.Now().Add(2 * 24 * time.Hour).UnixMilli(),
	}
	err = ts.transaction.Action(func(conn database.DbConn) error {
		err = ts.taskRepo.SaveTask(ctx, conn, ta)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTask error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		//增加指派人员
		tm := &data.TaskMember{
			MemberCode: assignTo,
			TaskCode:   ta.Id,
			JoinTime:   time.Now().UnixMilli(),
			IsOwner:    model.Owner,
		}
		if assignTo == msg.MemberId {
			tm.IsExecutor = model.Executor
		}
		err = ts.taskRepo.SaveTaskMember(ctx, conn, tm)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTaskMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 调用User grpc 端口查询对应用户信息
	display := ta.ToTaskDisplay()
	//member, err := rpc.UserGrpcClient.FindMemberById(ctx, &login.UserMessage{MemId: assignTo})
	//if err != nil {
	//	return nil, err
	//}
	member, err := ts.userRpcDomain.MemberInfo(ctx, assignTo)
	if err != nil {
		zap.L().Error("project task SaveTask ts.userRpcDomain.MemberInfo error", zap.Error(err))
		return nil, err
	}

	display.Executor = data.Executor{
		Name:   member.Name,
		Avatar: member.Avatar,
		Code:   member.Code,
	}
	tm := &task.TaskMessage{}
	copier.Copy(tm, display)

	// 添加任务动态=》 ms_project_log
	createProjectLog(ts.projectLogRepo, ta.ProjectCode, ta.Id, ta.Name, ta.AssignTo, "create", "task")

	// 保存的时候需要添加 kafka 缓存删除信息
	//发送kafka 缓存删除
	config.SendCache([]byte("task"))
	
	return tm, nil

}

// 辅助函数（添加任务动态）
func createProjectLog(
	logRepo repo.ProjectLogRepo,
	projectCode int64,
	taskCode int64,
	taskName string,
	toMemberCode int64,
	logType string,
	actionType string) {
	remark := ""
	if logType == "create" {
		remark = "创建了任务"
	}
	pl := &data.ProjectLog{
		MemberCode:  toMemberCode,
		SourceCode:  taskCode,
		Content:     taskName,
		Remark:      remark,
		ProjectCode: projectCode,
		CreateTime:  time.Now().UnixMilli(),
		Type:        logType,
		ActionType:  actionType,
		Icon:        "plus",
		IsComment:   0,
		IsRobot:     0,
	}
	logRepo.SaveProjectLog(pl)
}

func (ts *TaskService) TaskSort(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskSortResponse, error) {
	//移动的任务id肯定有 preTaskCode
	preTaskCode := encrypts.DecryptNoErr(msg.PreTaskCode)
	toStageCode := encrypts.DecryptNoErr(msg.ToStageCode)
	if msg.PreTaskCode == msg.NextTaskCode {
		return &task.TaskSortResponse{}, nil
	}
	err := ts.SortTask(ctx, preTaskCode, msg.NextTaskCode, toStageCode)
	if err != nil {
		return nil, err
	}
	return &task.TaskSortResponse{}, nil

}

// sort排序更新的方法 TaskSort => SortTask => resetSort
func (ts *TaskService) SortTask(ctx context.Context, preTaskCode int64, nextTaskCode string, toStageCode int64) error {
	//1. 从小到大排
	//2. 原有的顺序  比如 1 2 3 4 5 4排到2前面去 4的序号在1和2 之间 如果4是最后一个 保证 4比所有的序号都打 如果 排到第一位 直接置为0
	ta, err := ts.taskRepo.FindTaskById(ctx, preTaskCode)
	if err != nil {
		zap.L().Error("project task TaskSort taskRepo.FindTaskById error", zap.Error(err))
		return errs.GrpcError(model.DBError)
	}
	//如果相等是不需要进行改变的
	ta.StageCode = int(toStageCode)
	var nextTs *data.Task
	err = ts.transaction.Action(func(conn database.DbConn) error {
		if nextTaskCode != "" {
			//顺序变了 需要互换位置
			nextTaskId := encrypts.DecryptNoErr(nextTaskCode)
			next, err := ts.taskRepo.FindTaskById(ctx, nextTaskId)
			nextTs = next
			if err != nil {
				zap.L().Error("project task TaskSort nextTaskId taskRepo.FindTaskById error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			// next.Sort 要找到当前stages比它小的那个任务
			prepare, err := ts.taskRepo.FindTaskByStageCodeLtSort(ctx, next.StageCode, next.Sort)
			if err != nil {
				zap.L().Error("project task TaskSort taskRepo.FindTaskByStageCodeLtSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			if prepare != nil { // 取当前stages 前一个Sort 和 后一个 sort的中间值，为此，sort序列间需要加大间隔 65536，小于一定间隔需要重置 resetSort
				ta.Sort = (prepare.Sort + next.Sort) / 2
			}
			if prepare == nil {
				ta.Sort = 0
			}

		} else { // 没有后一位task的时候（说明在末尾）
			//找到当前stages最大的task sort
			maxSort, err := ts.taskRepo.FindTaskSort(ctx, ta.ProjectCode, int64(ta.StageCode))
			if err != nil {
				zap.L().Error("project task TaskSort taskRepo.FindTaskSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			if maxSort == nil { // 说明整个stage 没有task，这是第一个，需要赋予初始值
				a := 0
				maxSort = &a
			}
			ta.Sort = *maxSort + 65536
		}
		err = ts.taskRepo.UpdateTaskSort(ctx, conn, ta) // 更新移动后的task信息（stage，sort）
		if err != nil {
			zap.L().Error("project task TaskSort taskRepo.UpdateTaskSort error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	if err != nil {
		zap.L().Error("project task TaskSort ts.transaction.Action error", zap.Error(err))
		return errs.GrpcError(model.DBError)
	}
	if (ta.Sort < 50) || (nextTs != nil && math.Abs(float64(nextTs.Sort-ta.Sort)) < 50) { // 当前排序间隔小于50
		//重置排序
		err = ts.resetSort(toStageCode)
		if err != nil {
			zap.L().Error("project task TaskSort resetSort error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	}
	return err
}

// 重置sort的方法
func (ts *TaskService) resetSort(stageCode int64) error {
	list, err := ts.taskRepo.FindTaskByStageCode(context.Background(), int(stageCode))
	if err != nil {
		return err
	}
	return ts.transaction.Action(func(conn database.DbConn) error {
		iSort := 65536
		for index, v := range list {
			v.Sort = (index + 1) * iSort
			ts.taskRepo.UpdateTaskSort(context.Background(), conn, v)
		}
		return nil
	})
}

func (ts *TaskService) MyTaskList(ctx context.Context, msg *task.TaskReqMessage) (*task.MyTaskListResponse, error) {
	var tsList []*data.Task
	var err error
	var total int64
	switch msg.TaskType {
	case 1:
		//我执行的
		tsList, total, err = ts.taskRepo.FindTaskByAssignTo(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByAssignTo error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	case 2:
		//我参与的
		tsList, total, err = ts.taskRepo.FindTaskByMemberCode(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByMemberCode error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	case 3:
		//我创建的
		tsList, total, err = ts.taskRepo.FindTaskByCreateBy(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByCreateBy error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}

	if tsList == nil || len(tsList) <= 0 { // 如果为空就即刻返回
		return &task.MyTaskListResponse{List: nil, Total: 0}, nil
	}

	// 查询对应task 的 project && member信息
	var pids []int64
	var mids []int64
	for _, v := range tsList {
		pids = append(pids, v.ProjectCode)
		mids = append(mids, v.AssignTo)
	}

	// 协程并发查询project & member
	plistChan := make(chan []*data.Project)
	defer close(plistChan)
	pErrChan := make(chan error)
	defer close(pErrChan)

	mlistChan := make(chan *login.MemberMessageList)
	defer close(mlistChan)
	mErrChan := make(chan error)
	defer close(mErrChan)

	go func() {
		pList, err := ts.projectRepo.FindProjectByIds(ctx, pids)
		plistChan <- pList
		pErrChan <- err
	}()

	go func() {
		//mList, err := rpc.UserGrpcClient.FindMemInfoByIds(ctx, &login.UserMessage{
		//	MIds: mids,
		//})
		mList, _, err := ts.userRpcDomain.MemberList(ctx, mids)
		mlistChan <- mList
		mErrChan <- err
	}()

	pList := <-plistChan
	if err := <-pErrChan; err != nil {
		zap.L().Error("project task MyTaskList projectRepo.FindProjectByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	projectMap := data.ToProjectMap(pList)

	mList := <-mlistChan
	if err := <-mErrChan; err != nil {
		zap.L().Error("project task MyTaskList ts.userRpcDomain.MemberList error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range mList.List {
		mMap[v.Id] = v
	}
	var mtdList []*data.MyTaskDisplay
	for _, v := range tsList {
		memberMessage := mMap[v.AssignTo]
		name := memberMessage.Name
		avatar := memberMessage.Avatar
		mtd := v.ToMyTaskDisplay(projectMap[v.ProjectCode], name, avatar)
		mtdList = append(mtdList, mtd)
	}
	var myMsgs []*task.MyTaskMessage
	copier.Copy(&myMsgs, mtdList)
	return &task.MyTaskListResponse{List: myMsgs, Total: total}, nil
}

func (ts *TaskService) TaskRead(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskMessage, error) {
	//根据taskCode查询任务详情 根据任务查询项目详情 根据任务查询任务步骤详情 查询任务的执行者的成员详情
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	taskInfo, err := ts.taskRepo.FindTaskById(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task ReadTask taskRepo FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskInfo == nil {
		return &task.TaskMessage{}, nil
	}
	display := taskInfo.ToTaskDisplay()
	if taskInfo.Private == 1 {
		//代表隐私模式
		taskMember, err := ts.taskRepo.FindTaskMemberByTaskId(ctx, taskInfo.Id, msg.MemberId)
		if err != nil {
			zap.L().Error("project task TaskList taskRepo.FindTaskMemberByTaskId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		if taskMember != nil {
			display.CanRead = model.CanRead
		} else {
			display.CanRead = model.NoCanRead
		}
	}
	pj, err := ts.projectRepo.FindProjectById(ctx, taskInfo.ProjectCode)
	display.ProjectName = pj.Name
	taskStages, err := ts.taskStagesRepo.FindById(ctx, taskInfo.StageCode)
	display.StageName = taskStages.Name
	//memberMessage, err := rpc.UserGrpcClient.FindMemberById(ctx, &login.UserMessage{MemId: taskInfo.AssignTo})
	memberMessage, err := ts.userRpcDomain.MemberInfo(ctx, taskInfo.AssignTo)
	if err != nil {
		zap.L().Error("project task TaskList ts.userRpcDomain.MemberInfo error", zap.Error(err))
		return nil, err
	}
	e := data.Executor{
		Name:   memberMessage.Name,
		Avatar: memberMessage.Avatar,
	}
	display.Executor = e
	var taskMessage = &task.TaskMessage{}
	copier.Copy(taskMessage, display)
	return taskMessage, nil
}

func (ts *TaskService) ListTaskMember(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskMemberList, error) {
	//查询 task member表 根据memberCode去查询用户信息
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	taskMemberPage, total, err := ts.taskRepo.FindTaskMemberPage(ctx, taskCode, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project task TaskList taskRepo.FindTaskMemberPage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var mids []int64
	for _, v := range taskMemberPage {
		mids = append(mids, v.MemberCode)
	}
	//messageList, err := rpc.UserGrpcClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mids})
	//mMap := make(map[int64]*login.MemberMessage, len(messageList.List))
	//for _, v := range messageList.List {
	//	mMap[v.Id] = v
	//}
	_, mMap, err := ts.userRpcDomain.MemberList(ctx, mids)

	var taskMemeberMemssages []*task.TaskMemberMessage
	for _, v := range taskMemberPage {
		tm := &task.TaskMemberMessage{}
		tm.Code = encrypts.EncryptNoErr(v.MemberCode)
		tm.Id = v.Id
		message := mMap[v.MemberCode]
		tm.Name = message.Name
		tm.Avatar = message.Avatar
		tm.IsExecutor = int32(v.IsExecutor)
		tm.IsOwner = int32(v.IsOwner)
		taskMemeberMemssages = append(taskMemeberMemssages, tm)
	}
	return &task.TaskMemberList{List: taskMemeberMemssages, Total: total}, nil
}

func (ts *TaskService) TaskLog(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskLogList, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	all := msg.All
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	var list []*data.ProjectLog
	var total int64
	var err error
	if all == 1 {
		//显示全部
		list, total, err = ts.projectLogRepo.FindLogByTaskCode(ctx, taskCode, int(msg.Comment))
	}
	if all == 0 {
		//分页
		list, total, err = ts.projectLogRepo.FindLogByTaskCodePage(ctx, taskCode, int(msg.Comment), int(msg.Page), int(msg.PageSize))
	}
	if err != nil {
		zap.L().Error("project task TaskLog projectLogRepo.FindLogByTaskCodePage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if total == 0 {
		return &task.TaskLogList{}, nil
	}
	var displayList []*data.ProjectLogDisplay

	//调用user grpc 模块查找用户信息
	var mIdList []int64
	for _, v := range list {
		mIdList = append(mIdList, v.MemberCode)
	}
	//messageList, err := rpc.UserGrpcClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIdList})
	//mMap := make(map[int64]*login.MemberMessage)
	//for _, v := range messageList.List {
	//	mMap[v.Id] = v
	//}
	_, mMap, err := ts.userRpcDomain.MemberList(ctx, mIdList)

	for _, v := range list {
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := data.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}
	var l []*task.TaskLog
	copier.Copy(&l, displayList)
	return &task.TaskLogList{List: l, Total: total}, nil
}

func (ts *TaskService) TaskWorkTimeList(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskWorkTimeResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	/*var list []*data.TaskWorkTime
	var err error
	list, err = ts.taskWorkTimeRepo.FindWorkTimeList(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task TaskWorkTimeList taskWorkTimeRepo.FindWorkTimeList error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(list) == 0 {
		return &task.TaskWorkTimeResponse{}, nil
	}

	var displayList []*data.TaskWorkTimeDisplay

	//调用user grpc 查找member信息
	var mIdList []int64
	for _, v := range list {
		mIdList = append(mIdList, v.MemberCode)
	}
	//messageList, err := rpc.UserGrpcClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIdList})
	//mMap := make(map[int64]*login.MemberMessage)
	//for _, v := range messageList.List {
	//	mMap[v.Id] = v
	//}

	_, mMap, err := ts.userRpcDomain.MemberList(ctx, mIdList)
	for _, v := range list {
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := data.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}*/
	displayList, err := ts.taskWorkTimeDomain.TaskWorkTimeList(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task TaskWorkTimeList taskWorkTimeRepo.FindWorkTimeList error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var l []*task.TaskWorkTime
	copier.Copy(&l, displayList)
	return &task.TaskWorkTimeResponse{List: l, Total: int64(len(l))}, nil
}

func (ts *TaskService) SaveTaskWorkTime(ctx context.Context, msg *task.TaskReqMessage) (*task.SaveTaskWorkTimeResponse, error) {
	tmt := &data.TaskWorkTime{}
	tmt.BeginTime = msg.BeginTime
	tmt.Num = int(msg.Num)
	tmt.Content = msg.Content
	tmt.TaskCode = encrypts.DecryptNoErr(msg.TaskCode)
	tmt.MemberCode = msg.MemberId
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	err := ts.taskWorkTimeDomain.Save(ctx, tmt)
	if err != nil {
		zap.L().Error("project task SaveTaskWorkTime taskWorkTimeRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &task.SaveTaskWorkTimeResponse{}, nil
}

func (ts *TaskService) SaveTaskFile(ctx context.Context, msg *task.TaskFileReqMessage) (*task.TaskFileResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	//存file表
	f := &data.File{
		PathName:         msg.PathName,
		Title:            msg.FileName,
		Extension:        msg.Extension,
		Size:             int(msg.Size),
		ObjectType:       "",
		OrganizationCode: encrypts.DecryptNoErr(msg.OrganizationCode),
		TaskCode:         encrypts.DecryptNoErr(msg.TaskCode),
		ProjectCode:      encrypts.DecryptNoErr(msg.ProjectCode),
		CreateBy:         msg.MemberId,
		CreateTime:       time.Now().UnixMilli(),
		Downloads:        0,
		Extra:            "",
		Deleted:          model.NoDeleted,
		FileType:         msg.FileType,
		FileUrl:          msg.FileUrl,
		DeletedTime:      0,
	}
	err := ts.fileRepo.Save(ctx, f)
	if err != nil {
		zap.L().Error("project task SaveTaskFile fileRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//存入source_link
	sl := &data.SourceLink{
		SourceType:       "file",
		SourceCode:       f.Id,
		LinkType:         "task",
		LinkCode:         taskCode,
		OrganizationCode: encrypts.DecryptNoErr(msg.OrganizationCode),
		CreateBy:         msg.MemberId,
		CreateTime:       time.Now().UnixMilli(),
		Sort:             0,
	}
	err = ts.sourceLinkRepo.Save(context.Background(), sl)
	if err != nil {
		zap.L().Error("project task SaveTaskFile sourceLinkRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &task.TaskFileResponse{}, nil
}

func (ts *TaskService) TaskSources(ctx context.Context, msg *task.TaskReqMessage) (*task.TaskSourceResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	// source_link 表中查询对应 task 的文件 id
	sourceLinks, err := ts.sourceLinkRepo.FindByTaskCode(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task SaveTaskFile sourceLinkRepo.FindByTaskCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(sourceLinks) == 0 {
		return &task.TaskSourceResponse{}, nil
	}
	// 根据文件Id 查询 文件详细信息
	var fIdList []int64
	for _, v := range sourceLinks {
		fIdList = append(fIdList, v.SourceCode)
	}
	files, err := ts.fileRepo.FindByIds(context.Background(), fIdList)
	if err != nil {
		zap.L().Error("project task SaveTaskFile fileRepo.FindByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	//拼装数据
	fMap := make(map[int64]*data.File)
	for _, v := range files {
		fMap[v.Id] = v
	}
	var list []*data.SourceLinkDisplay
	for _, v := range sourceLinks {
		list = append(list, v.ToDisplay(fMap[v.SourceCode]))
	}
	var slMsg []*task.TaskSourceMessage
	copier.Copy(&slMsg, list)
	return &task.TaskSourceResponse{List: slMsg}, nil
}

func (ts *TaskService) CreateComment(ctx context.Context, msg *task.TaskReqMessage) (*task.CreateCommentResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	taskById, err := ts.taskRepo.FindTaskById(ctx, taskCode)
	if err != nil {
		zap.L().Error("project task CreateComment fileRepo.FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pl := &data.ProjectLog{
		MemberCode:   msg.MemberId,
		Content:      msg.CommentContent,
		Remark:       msg.CommentContent,
		Type:         "createComment",
		CreateTime:   time.Now().UnixMilli(),
		SourceCode:   taskCode,
		ActionType:   "task",
		ToMemberCode: 0,
		IsComment:    model.Comment,
		ProjectCode:  taskById.ProjectCode,
		Icon:         "plus",
		IsRobot:      0,
	}
	ts.projectLogRepo.SaveProjectLog(pl)
	return &task.CreateCommentResponse{}, nil
}
