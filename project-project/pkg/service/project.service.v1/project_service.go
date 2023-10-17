package project_service_v1

import (
	context "context"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"project-common/errs"
	"project-grpc/project"
	"project-project/internal/dao"
	"project-project/internal/data/menu"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
	"project-project/pkg/model"
)

type ProjectService struct {
	project.UnimplementedProjectServiceServer
	cache       repo.Cache
	transaction tran.Transaction
	menuRepo    repo.MenuRepo
}

func New() *ProjectService {
	return &ProjectService{
		cache:       dao.Rc,
		transaction: dao.NewTransactionDao(),
		menuRepo:    dao.NewMenuDao(),
	}
}

func (ps *ProjectService) Index(ctx context.Context, msg *project.IndexMessage) (*project.IndexResponse, error) {
	c := context.Background()
	//1. index 获取所有的menu页资料
	pms, err := ps.menuRepo.FindMenus(c)
	if err != nil {
		zap.L().Error("Project db FindMenus error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if pms == nil {
		return nil, errs.GrpcError(model.SyntaxError)
	}
	//2. 构建IndexResponse的MenuMessage递归树
	var mms []*project.MenuMessage
	childTrees := menu.CovertChild(pms)
	copier.Copy(&mms, &childTrees)
	// 回复grpc响应
	return &project.IndexResponse{Menus: mms}, nil
}
