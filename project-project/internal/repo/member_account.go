package repo

import (
	"context"
	"project-project/internal/data"
)

type MemberAccountRepo interface {
	FindList(ctx context.Context, condition string, organizationCode int64, departmentCode int64, page int64, pageSize int64) ([]*data.MemberAccount, int64, error)
	FindByMemberId(background context.Context, memberId int64) (*data.MemberAccount, error)
	Save(ctx context.Context, memberAccount *data.MemberAccount) error
}
