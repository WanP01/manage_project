package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type MemberAccountDao struct {
	conn *gorms.GormConn
}

func (m *MemberAccountDao) Save(ctx context.Context, memberAccount *data.MemberAccount) error {
	err := m.conn.Session(ctx).Save(memberAccount).Error
	return err
}

func NewMemberAccountDao() *MemberAccountDao {
	return &MemberAccountDao{
		conn: gorms.New(),
	}
}

func (m *MemberAccountDao) FindByMemberId(ctx context.Context, memberId int64) (*data.MemberAccount, error) {
	var ma *data.MemberAccount
	var err error
	session := m.conn.Session(ctx)
	err = session.Where("member_code=?", memberId).Take(&ma).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return ma, err
}

func (m *MemberAccountDao) FindList(ctx context.Context, condition string, organizationCode int64, departmentCode int64, page int64, pageSize int64) ([]*data.MemberAccount, int64, error) {
	var list []*data.MemberAccount
	var total int64
	var err error
	session := m.conn.Session(ctx)
	offset := (page - 1) * pageSize
	err = session.Model(&data.MemberAccount{}).
		Where("organization_code=?", organizationCode).
		Where(condition).Limit(int(pageSize)).Offset(int(offset)).Find(&list).Error
	err = session.Model(&data.MemberAccount{}).
		Where("organization_code=?", organizationCode).
		Where(condition).Count(&total).Error
	return list, total, err
}
