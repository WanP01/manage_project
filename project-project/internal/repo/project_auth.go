package repo

import (
	"context"
	"project-project/internal/data"
)

type ProjectAuthRepo interface {
	FindAuthList(ctx context.Context, orgCode int64) ([]*data.ProjectAuth, error)
	FindAuthListPage(ctx context.Context, orgCode int64, page int64, pageSize int64) ([]*data.ProjectAuth, int64, error)
	FindAuthListNoOrg(ctx context.Context) ([]*data.ProjectAuth, error)
	Save(ctx context.Context, pa *data.ProjectAuth) error
	FindAuthByTitleAndOrgCode(ctx context.Context, title string, orgCode int64) (*data.ProjectAuth, error)
}
