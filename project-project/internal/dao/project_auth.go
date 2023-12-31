package dao

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type ProjectAuthDao struct {
	conn *gorms.GormConn
}

func (p *ProjectAuthDao) FindAuthByTitleAndOrgCode(ctx context.Context, title string, orgCode int64) (*data.ProjectAuth, error) {
	var prA *data.ProjectAuth
	err := p.conn.Session(ctx).Model(&data.ProjectAuth{}).Where("title = ? and organization_code = ?", title, orgCode).First(&prA).Error
	return prA, err
}

func (p *ProjectAuthDao) Save(ctx context.Context, pa *data.ProjectAuth) error {
	err := p.conn.Session(ctx).Save(&pa).Error
	return err
}

func NewProjectAuthDao() *ProjectAuthDao {
	return &ProjectAuthDao{
		conn: gorms.New(),
	}
}

func (p *ProjectAuthDao) FindAuthListPage(ctx context.Context, orgCode int64, page int64, pageSize int64) ([]*data.ProjectAuth, int64, error) {

	var list []*data.ProjectAuth
	var total int64
	var err error
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectAuth{}).
		Where("organization_code=?", orgCode).
		Limit(int(pageSize)).
		Offset(int((page - 1) * pageSize)).
		Find(&list).Error
	err = session.Model(&data.ProjectAuth{}).
		Where("organization_code=?", orgCode).
		Count(&total).Error
	return list, total, err
}

func (p *ProjectAuthDao) FindAuthList(ctx context.Context, orgCode int64) ([]*data.ProjectAuth, error) {
	var list []*data.ProjectAuth
	var err error
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectAuth{}).Where("organization_code=?", orgCode).Find(&list).Error
	return list, err
}

func (p *ProjectAuthDao) FindAuthListNoOrg(ctx context.Context) ([]*data.ProjectAuth, error) {
	var list []*data.ProjectAuth
	var err error
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectAuth{}).Where("organization_code IS NUll").Find(&list).Error
	return list, err
}
