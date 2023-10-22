package repo

import (
	"context"
	"project-project/internal/data"
)

type TaskWorkTimeRepo interface {
	Save(ctx context.Context, twt *data.TaskWorkTime) error
	FindWorkTimeList(ctx context.Context, taskCode int64) ([]*data.TaskWorkTime, error)
}
