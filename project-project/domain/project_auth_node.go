package domain

import (
	"context"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/database"
	"project-project/internal/repo"
	"project-project/pkg/model"
)

type ProjectAuthNodeDomain struct {
	projectAuthNodeRepo repo.ProjectAuthNodeRepo
}

func NewProjectAuthNodeDomain() *ProjectAuthNodeDomain {
	return &ProjectAuthNodeDomain{
		projectAuthNodeRepo: dao.NewProjectAuthNodeDao(),
	}
}

func (d *ProjectAuthNodeDomain) AuthNodeList(ctx context.Context, authId int64) ([]string, *errs.BError) {
	list, err := d.projectAuthNodeRepo.FindNodeStringList(ctx, authId)
	if err != nil {
		return nil, model.DBError
	}
	return list, nil
}

func (d *ProjectAuthNodeDomain) Save(ctx context.Context, conn database.DbConn, authId int64, nodes []string) *errs.BError {
	err := d.projectAuthNodeRepo.DeleteByAuthId(ctx, conn, authId)
	if err != nil {
		return model.DBError
	}
	err = d.projectAuthNodeRepo.Save(ctx, conn, authId, nodes)
	if err != nil {
		return model.DBError
	}
	return nil
}
