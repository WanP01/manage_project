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

type ProjectAuthDomain struct {
	projectAuthRepo repo.ProjectAuthRepo
}

func NewProjectAuthDomain() *ProjectAuthDomain {
	return &ProjectAuthDomain{
		projectAuthRepo: dao.NewProjectAuthDao(),
	}
}

func (d *ProjectAuthDomain) AuthList(ctx context.Context, orgCode int64) ([]*data.ProjectAuthDisplay, *errs.BError) {
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	list, err := d.projectAuthRepo.FindAuthList(ctx, orgCode)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, nil
}

func (d *ProjectAuthDomain) AuthListPage(ctx context.Context, orgCode int64, page int64, pageSize int64) ([]*data.ProjectAuthDisplay, int64, *errs.BError) {
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	list, total, err := d.projectAuthRepo.FindAuthListPage(ctx, orgCode, page, pageSize)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, 0, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, total, nil
}
