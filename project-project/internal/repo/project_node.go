package repo

import (
	"context"
	"project-project/internal/data"
)

type ProjectNodeRepo interface {
	FindAll(ctx context.Context) ([]*data.ProjectNode, error)
}
