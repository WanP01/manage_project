package repo

import (
	"context"
	"project-user/internal/data/member"
	"project-user/internal/database"
)

type MemberRepo interface {
	GetMemberByEmail(ctx context.Context, email string) (bool, error)
	GetMemBerByAccount(ctx context.Context, account string) (bool, error)
	GetMemberByMobile(ctx context.Context, mobile string) (bool, error)
	SaveMember(conn database.DbConn, ctx context.Context, mem *member.Member) error
	FindMember(ctx context.Context, account string, pwd string) (*member.Member, error)
}
