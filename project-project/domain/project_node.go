package domain

import (
	"context"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/repo"
	"project-project/pkg/model"
)

type ProjectNodeDomain struct {
	projectNodeRepo repo.ProjectNodeRepo
}

func (d *ProjectNodeDomain) TreeList(ctx context.Context) ([]*data.ProjectNodeTree, *errs.BError) {
	//node表都查出来 转换成treelist结构
	list, err := d.projectNodeRepo.FindAll(ctx)
	if err != nil {
		return nil, model.DBError
	}
	return data.ToNodeTreeList(list), nil
}

func (d *ProjectNodeDomain) NodeList(ctx context.Context) ([]*data.ProjectNode, *errs.BError) {
	list, err := d.projectNodeRepo.FindAll(ctx)
	if err != nil {
		return nil, model.DBError
	}
	return list, nil
}

func NewProjectNodeDomain() *ProjectNodeDomain {
	return &ProjectNodeDomain{
		projectNodeRepo: dao.NewProjectNodeDao(),
	}
}
