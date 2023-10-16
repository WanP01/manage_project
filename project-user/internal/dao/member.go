package dao

import (
	"context"
	"gorm.io/gorm"
	"project-user/internal/data/member"
	"project-user/internal/database"
	"project-user/internal/database/gorms"
)

type MemberDao struct {
	conn *gorms.GormConn
}

func NewMemberDao() *MemberDao {
	return &MemberDao{
		conn: gorms.New(),
	}
}

func (m MemberDao) GetMemberByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := m.conn.Session(ctx).Model(&member.Member{}).Where("email=?", email).Count(&count).Error
	return count > 0, err
}

func (m MemberDao) GetMemBerByAccount(ctx context.Context, account string) (bool, error) {
	var count int64
	err := m.conn.Session(ctx).Model(&member.Member{}).Where("account=?", account).Count(&count).Error
	return count > 0, err
}

func (m MemberDao) GetMemberByMobile(ctx context.Context, mobile string) (bool, error) {
	var count int64
	err := m.conn.Session(ctx).Model(&member.Member{}).Where("mobile=?", mobile).Count(&count).Error
	return count > 0, err
}

func (m MemberDao) SaveMember(conn database.DbConn, ctx context.Context, mem *member.Member) error {
	m.conn = conn.(*gorms.GormConn)
	err := m.conn.Tx(ctx).Create(mem).Error
	return err
}

func (m MemberDao) FindMember(ctx context.Context, account string, pwd string) (*member.Member, error) {
	var meminfo *member.Member
	err := m.conn.Session(ctx).Where("account=? and password=?", account, pwd).First(&meminfo).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return meminfo, err
}
