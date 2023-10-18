package dao

import (
	"context"
	"project-project/internal/data/pro"
	"project-project/internal/database/gorms"
)

type ProjectTemplateDao struct {
	conn *gorms.GormConn
}

func NewProjectTemplateDao() *ProjectTemplateDao {
	return &ProjectTemplateDao{
		conn: gorms.New(),
	}
}

// FindProjectTemplateSystem 查找系统模板列表
func (p *ProjectTemplateDao) FindProjectTemplateSystem(ctx context.Context, page int64, size int64) ([]pro.ProjectTemplate, int64, error) {
	var ptm []pro.ProjectTemplate
	err := p.conn.Session(ctx).Model(&pro.ProjectTemplate{}).
		Where("is_system=?", 1).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Find(&ptm).Error
	if err != nil {
		return ptm, 0, err
	}
	var total int64
	err = p.conn.Session(ctx).Model(&pro.ProjectTemplate{}).
		Where("is_system=?", 1).
		Count(&total).Error
	return ptm, total, err
}

// FindProjectTemplateCustom 查找用户模板列表（自定义模板）
func (p *ProjectTemplateDao) FindProjectTemplateCustom(ctx context.Context, memId int64, organizationCode int64, page int64, size int64) ([]pro.ProjectTemplate, int64, error) {
	var ptm []pro.ProjectTemplate
	err := p.conn.Session(ctx).Model(&pro.ProjectTemplate{}).
		Where("is_system=? and member_code=? and organization_code=?", 0, memId, organizationCode).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Find(&ptm).Error
	if err != nil {
		return ptm, 0, err
	}
	var total int64
	err = p.conn.Session(ctx).Model(&pro.ProjectTemplate{}).
		Where("is_system=? and member_code=? and organization_code=?", 0, memId, organizationCode).
		Count(&total).Error
	return ptm, total, err
}

// FindProjectTemplateAll 查找组织模板列表
func (p *ProjectTemplateDao) FindProjectTemplateAll(ctx context.Context, organizationCode int64, page int64, size int64) ([]pro.ProjectTemplate, int64, error) {
	var ptm []pro.ProjectTemplate
	err := p.conn.Session(ctx).Model(&pro.ProjectTemplate{}).
		Where("organization_code=?", organizationCode).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Find(&ptm).Error
	if err != nil {
		return ptm, 0, err
	}
	var total int64
	err = p.conn.Session(ctx).Model(&pro.ProjectTemplate{}).
		Where("organization_code=?", organizationCode).
		Count(&total).Error
	return ptm, total, err
}