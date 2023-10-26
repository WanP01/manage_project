package repo

import (
	"context"
	"project-project/internal/database"
)

type ProjectAuthNodeRepo interface {
	FindNodeStringList(ctx context.Context, authId int64) ([]string, error)
	DeleteByAuthId(ctx context.Context, conn database.DbConn, authId int64) error
	Save(ctx context.Context, conn database.DbConn, authId int64, nodes []string) error
}
