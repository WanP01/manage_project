package repo

import (
	"context"
	"project-project/internal/data"
)

type ProjectAuthRepo interface {
	FindAuthList(ctx context.Context, orgCode int64) ([]*data.ProjectAuth, error)
	FindAuthListPage(ctx context.Context, orgCode int64, page int64, pageSize int64) ([]*data.ProjectAuth, int64, error)
}
