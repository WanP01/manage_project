package repo

import (
	"context"
	"project-project/internal/data"
)

type FileRepo interface {
	Save(ctx context.Context, file *data.File) error
	FindByIds(background context.Context, ids []int64) ([]*data.File, error)
}
