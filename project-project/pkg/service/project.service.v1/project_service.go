package project_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"project-common/encrypts"
	"project-common/errs"
	"project-common/tms"
	"project-grpc/project"
	"project-grpc/user/login"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
	"project-project/internal/rpc"
	"project-project/pkg/model"
	"strconv"
	"time"
)

type ProjectService struct {
	project.UnimplementedProjectServiceServer
	cache                  repo.Cache
	transaction            tran.Transaction
	menuRepo               repo.MenuRepo
	projectRepo            repo.ProjectRepo
	projectTemplateRepo    repo.ProjectTemplateRepo
	taskStagesTemplateRepo repo.TaskStagesTemplateRepo
	taskStagesRepo         repo.TaskStagesRepo
}

func New() *ProjectService {
	return &ProjectService{
		cache:                  dao.Rc,
		transaction:            dao.NewTransactionDao(),
		menuRepo:               dao.NewMenuDao(),
		projectRepo:            dao.NewProjectDao(),
		projectTemplateRepo:    dao.NewProjectTemplateDao(),
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(),
		taskStagesRepo:         dao.NewTaskStagesDao(),
	}
}

func (ps *ProjectService) Index(ctx context.Context, msg *project.IndexMessage) (*project.IndexResponse, error) {

	//1. index 获取所有的menu页资料
	pms, err := ps.menuRepo.FindMenus(ctx)
	if err != nil {
		zap.L().Error("Project db FindMenus error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if pms == nil {
		return nil, errs.GrpcError(model.SyntaxError)
	}
	//2. 构建IndexResponse的MenuMessage递归树
	var mms []*project.MenuMessage
	childTrees := data.CovertChild(pms)
	copier.Copy(&mms, &childTrees)
	// 回复grpc响应
	return &project.IndexResponse{Menus: mms}, nil
}

func (ps *ProjectService) FindProjectByMemId(ctx context.Context, msg *project.ProjectRpcMessage) (*project.MyProjectResponse, error) {
	memId := msg.MemberId
	page := msg.Page
	pageSzie := msg.PageSize
	var pms []*data.ProjectAndMember
	var total int64
	var err error
	switch selectBy := msg.SelectBy; selectBy {
	case "", "my":
		pms, total, err = ps.projectRepo.FindProjectByMemID(ctx, "and deleted=0", memId, page, pageSzie)
	case "archive":
		pms, total, err = ps.projectRepo.FindProjectByMemID(ctx, "and archive=1 and deleted=0", memId, page, pageSzie)
	case "deleted":
		pms, total, err = ps.projectRepo.FindProjectByMemID(ctx, "and deleted=1", memId, page, pageSzie)
	case "collect":
		pms, total, err = ps.projectRepo.FindCollectProjectByMemID(ctx, "and deleted=0", memId, page, pageSzie)
	}
	//刷新收藏状态（可能仅部分行有）
	collectPms, _, err := ps.projectRepo.FindCollectProjectByMemID(ctx, "and deleted=0", memId, page, pageSzie)
	if err != nil {
		zap.L().Error("project FindProjectByMemId::FindCollectProjectByMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	cMap := make(map[int64]*data.ProjectAndMember)
	for _, c := range collectPms {
		cMap[c.Id] = c
	}
	for _, v := range pms {
		if cMap[v.ProjectCode] != nil {
			v.Collected = model.Collected
		}
	}
	//if pms == nil { // 返回默认值（空的Project）
	//	return &project.MyProjectResponse{Pm: []*project.ProjectMessage{&project.ProjectMessage{Name: "None", Description: "尚未创建"}}, Total: total}, nil
	//}
	if err != nil {
		zap.L().Error("project FindProjectByMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if pms == nil {
		return &project.MyProjectResponse{Pm: []*project.ProjectMessage{}, Total: total}, nil
	}
	var pmm []*project.ProjectMessage
	copier.Copy(&pmm, pms) // 空Pms对应pmm仍为nil，在api处再次赋值 空 []
	for _, v := range pmm {
		v.Code, _ = encrypts.EncryptInt64(v.ProjectCode, model.AESKey)
		pam := data.ToMap(pms)[v.Id]
		// 格式转换 int（数据库）=》string（前端） & 赋值
		v.AccessControlType = pam.GetAccessControlType()
		v.OrganizationCode, _ = encrypts.EncryptInt64(pam.OrganizationCode, model.AESKey)
		v.JoinTime = tms.FormatByMill(pam.JoinTime)
		v.OwnerName = msg.MemberName
		v.Order = int32(pam.Sort)
		v.CreateTime = tms.FormatByMill(pam.CreateTime)
	}
	return &project.MyProjectResponse{Pm: pmm, Total: total}, nil
}

func (ps *ProjectService) FindProjectTemplate(ctx context.Context, msg *project.ProjectRpcMessage) (*project.ProjectTemplateResponse, error) {
	memId := msg.MemberId
	organizationCodeStr, _ := encrypts.Decrypt(msg.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(organizationCodeStr, 10, 64)
	page := msg.Page
	size := msg.PageSize
	var pts []data.ProjectTemplate
	var total int64
	var err error
	// 1. 根据 view type 去查询项目模板列表 template list
	switch viewtype := msg.ViewType; viewtype {
	case 0:
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateCustom(ctx, memId, organizationCode, page, size)
	case 1:
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateSystem(ctx, page, size)
	case -1:
		pts, total, err = ps.projectTemplateRepo.FindProjectTemplateAll(ctx, organizationCode, page, size)
	}
	if err != nil {
		zap.L().Error("project FindProjectTemplate FindProjectTemplateSystem error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 2. 模型转换，拿到模板id 取任务步骤模板表去查询
	ids := data.ToProjectTemplateIds(pts)
	tst, err := ps.taskStagesTemplateRepo.FindInProTemIds(ctx, ids)
	if err != nil {
		zap.L().Error("project FindProjectTemplate FindInProTemIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 3. 组装数据 database 模型转换为 grpc 数据结构（一般与最终传输的数据结构一致）
	var ptas []*data.ProjectTemplateAll
	for _, v := range pts {
		ptas = append(ptas, v.Convert(data.CovertProjectMap(tst)[v.Id]))
	}
	var pmMsgs []*project.ProjectTemplateMessage
	copier.Copy(&pmMsgs, ptas)
	return &project.ProjectTemplateResponse{Ptm: pmMsgs, Total: total}, nil
}

func (ps *ProjectService) SaveProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.SaveProjectMessage, error) {
	//取出信息并解密
	organizationCodeStr, _ := encrypts.Decrypt(msg.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(organizationCodeStr, 10, 64)
	templateCodeStr, _ := encrypts.Decrypt(msg.TemplateCode, model.AESKey)
	templateCode, _ := strconv.ParseInt(templateCodeStr, 10, 64)
	//获取模板信息（通过模板Id 查找 模板对应的 task，并保存project与task的关系=》ms_task_stages）
	stageTemplateList, err := ps.taskStagesTemplateRepo.FindByProjectTemplateId(ctx, int(templateCode))
	if err != nil {
		zap.L().Error("project SaveProject taskStagesTemplateRepo.FindByProjectTemplateId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pr := &data.Project{
		Name:              msg.Name,
		Description:       msg.Description,
		TemplateCode:      templateCode,
		CreateTime:        time.Now().UnixMilli(),
		Cover:             "https://img2.baidu.com/it/u=792555388,2449797505&fm=253&fmt=auto&app=138&f=JPEG?w=667&h=500",
		Deleted:           model.NoDeleted,
		Archive:           model.NoArchive,
		OrganizationCode:  organizationCode,
		AccessControlType: model.Open,
		TaskBoardTheme:    model.Simple,
	}
	err = ps.transaction.Action(func(conn database.DbConn) error {
		err := ps.projectRepo.SaveProject(conn, ctx, pr)
		if err != nil {
			zap.L().Error("project SaveProject SaveProject error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}

		//2. 保存项目和成员的关联表
		pm := &data.ProjectMember{
			ProjectCode: pr.Id,
			MemberCode:  msg.MemberId,
			JoinTime:    time.Now().UnixMilli(),
			IsOwner:     msg.MemberId,
			Authorize:   "",
		}
		err = ps.projectRepo.SaveProjectMember(conn, ctx, pm)
		if err != nil {
			zap.L().Error("project SaveProject SaveProjectMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		// 3. 保存task关系（生成任务步骤） project=》template=》task
		for index, v := range stageTemplateList {
			taskStage := &data.TaskStages{
				ProjectCode: pr.Id,
				Name:        v.Name,
				Sort:        index + 1,
				Description: "",
				CreateTime:  time.Now().UnixMilli(),
				Deleted:     model.NoDeleted,
			}
			err = ps.taskStagesRepo.SaveTaskStages(ctx, conn, taskStage)
			if err != nil {
				zap.L().Error("project SaveProject taskStagesRepo.SaveTaskStages error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	code, _ := encrypts.EncryptInt64(pr.Id, model.AESKey)
	rsp := &project.SaveProjectMessage{
		Id:               pr.Id,
		Code:             code,
		OrganizationCode: organizationCodeStr,
		Name:             pr.Name,
		Cover:            pr.Cover,
		CreateTime:       tms.FormatByMill(pr.CreateTime),
		TaskBoardTheme:   pr.TaskBoardTheme,
	}
	return rsp, nil
}

func (ps *ProjectService) FindProjectDetail(ctx context.Context, msg *project.ProjectRpcMessage) (*project.ProjectDetailMessage, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectId, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	memberId := msg.MemberId
	// 根据用户Id 和 project Id 查项目表
	projectAndMember, err := ps.projectRepo.FindProjectByPIDANDMemID(ctx, memberId, projectId)
	if err != nil {
		zap.L().Error("project FindProjectDetail FindProjectByPIDANDMemID error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//获得所有者信息（所有者Id ）
	ownerId := projectAndMember.IsOwner
	//与User模块互动查找用户信息（ownerName，ownerAvatar）
	meminfo, err := rpc.UserGrpcClient.FindMemberById(ctx, &login.UserMessage{MemId: ownerId})
	if err != nil {
		zap.L().Error("project  FindProjectDetail rpc.UserGrpcClient.FindMemberById", zap.Error(err))
		return nil, err
	}
	//是否被收藏
	isCollect, err := ps.projectRepo.FindCollectProjectByPIDANDMemID(ctx, memberId, projectId)
	if err != nil {
		zap.L().Error("project  FindProjectDetail  FindCollectProjectByPIDANDMemID error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if isCollect {
		projectAndMember.Collected = model.Collected
	}
	detail := &project.ProjectDetailMessage{}
	copier.Copy(&detail, projectAndMember)
	detail.OwnerAvatar = meminfo.Avatar
	detail.OwnerName = meminfo.Name
	detail.Code, _ = encrypts.EncryptInt64(projectAndMember.ProjectCode, model.AESKey)
	detail.AccessControlType = projectAndMember.GetAccessControlType()
	detail.OrganizationCode, _ = encrypts.EncryptInt64(projectAndMember.OrganizationCode, model.AESKey)
	detail.Order = int32(projectAndMember.Sort)
	detail.CreateTime = tms.FormatByMill(projectAndMember.CreateTime)
	return detail, nil
}

func (ps *ProjectService) UpdateDeletedProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.DeletedProjectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	err := ps.projectRepo.UpdateDeletedProject(ctx, projectCode, msg.Deleted)
	if err != nil {
		zap.L().Error("project RecycleProject DeleteProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project.DeletedProjectResponse{}, nil
}

func (ps *ProjectService) UpdateProject(ctx context.Context, msg *project.UpdateProjectMessage) (*project.UpdateProjectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	proj := &data.Project{
		Id:                 projectCode,
		Name:               msg.Name,
		Description:        msg.Description,
		Cover:              msg.Cover,
		TaskBoardTheme:     msg.TaskBoardTheme,
		Prefix:             msg.Prefix,
		Private:            int(msg.Private),
		OpenPrefix:         int(msg.OpenPrefix),
		OpenBeginTime:      int(msg.OpenBeginTime),
		OpenTaskPrivate:    int(msg.OpenTaskPrivate),
		Schedule:           msg.Schedule,
		AutoUpdateSchedule: int(msg.AutoUpdateSchedule),
	}
	err := ps.projectRepo.UpdateProject(ctx, proj)
	if err != nil {
		zap.L().Error("project UpdateProject::UpdateProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project.UpdateProjectResponse{}, nil
}
