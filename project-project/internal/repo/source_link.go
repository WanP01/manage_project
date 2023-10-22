package repo

import (
	"context"
	"project-project/internal/data"
)

type SourceLinkRepo interface {
	Save(ctx context.Context, link *data.SourceLink) error
	FindByTaskCode(ctx context.Context, taskCode int64) ([]*data.SourceLink, error)
}
