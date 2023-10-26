package domain

import (
	"context"
	"go.uber.org/zap"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/repo"
	"project-project/pkg/model"
)

type MenuDomain struct {
	menuRepo repo.MenuRepo
}

func NewMenuDomain() *MenuDomain {
	return &MenuDomain{
		menuRepo: dao.NewMenuDao(),
	}
}

func (md *MenuDomain) MenuList(ctx context.Context) ([]*data.ProjectMenuChild, *errs.BError) {
	pms, err := md.menuRepo.FindMenus(ctx)
	if err != nil {
		zap.L().Error("menu MenuDomain db FindMenus error", zap.Error(err))
		return nil, model.DBError
	}
	if pms == nil {
		return nil, nil
	}
	//2. 构建IndexResponse的MenuMessage递归树
	childTrees := data.CovertChild(pms)
	// 回复grpc响应
	return childTrees, nil
}
