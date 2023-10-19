package repo

import (
	"context"
	"project-project/internal/data/pro"
	"project-project/internal/database"
)

type ProjectRepo interface {
	FindProjectByMemID(ctx context.Context, condition string, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error)
	FindCollectProjectByMemID(ctx context.Context, condition string, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error)
	SaveProject(conn database.DbConn, ctx context.Context, pr *pro.Project) error
	SaveProjectMember(conn database.DbConn, ctx context.Context, pm *pro.ProjectMember) error
	FindProjectByPIDANDMemID(ctx context.Context, memId int64, pId int64) (*pro.ProjectAndMember, error)
	FindCollectProjectByPIDANDMemID(ctx context.Context, memId int64, pId int64) (bool, error)
	UpdateDeletedProject(ctx context.Context, projectCode int64, deleted bool) error
	SaveProjectCollect(ctx context.Context, pc *pro.ProjectCollection) error
	DeleteProjectCollect(ctx context.Context, memId int64, projectCode int64) error
	UpdateProject(ctx context.Context, pj *pro.Project) error
}

type ProjectTemplateRepo interface {
	FindProjectTemplateSystem(ctx context.Context, page int64, size int64) ([]pro.ProjectTemplate, int64, error)
	FindProjectTemplateCustom(ctx context.Context, memId int64, organizationCode int64, page int64, size int64) ([]pro.ProjectTemplate, int64, error)
	FindProjectTemplateAll(ctx context.Context, organizationCode int64, page int64, size int64) ([]pro.ProjectTemplate, int64, error)
}
