package repo

import (
	"context"
	"project-project/internal/data"
)

type ProjectLogRepo interface {
	FindLogByTaskCode(ctx context.Context, taskCode int64, comment int) ([]*data.ProjectLog, int64, error)
	FindLogByTaskCodePage(ctx context.Context, taskCode int64, comment int, page int, pageSize int) ([]*data.ProjectLog, int64, error)
	SaveProjectLog(pl *data.ProjectLog)
	FindLogByMemberCode(background context.Context, memberId int64, page int64, size int64) ([]*data.ProjectLog, int64, error)
}
