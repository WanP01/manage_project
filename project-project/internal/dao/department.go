package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type DepartmentDao struct {
	conn *gorms.GormConn
}

func (d *DepartmentDao) Save(ctx context.Context, dpm *data.Department) error {
	err := d.conn.Session(ctx).Save(&dpm).Error
	return err
}

func (d *DepartmentDao) FindDepartment(ctx context.Context, organizationCode int64, parentDepartmentCode int64, name string) (*data.Department, error) {
	session := d.conn.Session(ctx)
	session = session.Model(&data.Department{}).Where("organization_code=? AND name=?", organizationCode, name)
	if parentDepartmentCode > 0 {
		session = session.Where("pcode=?", parentDepartmentCode)
	}
	var dp *data.Department
	err := session.Limit(1).Take(&dp).Error
	//判断是否为空，因为dp查不到的情况下会被赋值默认值，且err不为空
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return dp, err
}

func (d *DepartmentDao) ListDepartment(ctx context.Context, organizationCode int64, parentDepartmentCode int64, page int64, size int64) ([]*data.Department, int64, error) {
	var list []*data.Department
	var total int64
	var err error
	session := d.conn.Session(ctx)
	session = session.Model(&data.Department{})
	session = session.Where("organization_code=?", organizationCode)
	if parentDepartmentCode > 0 {
		session = session.Where("pcode=?", parentDepartmentCode)
	}
	err = session.Count(&total).Error
	err = session.Limit(int(size)).Offset(int((page - 1) * size)).Find(&list).Error
	return list, total, err
}

func (d *DepartmentDao) FindDepartmentById(ctx context.Context, id int64) (*data.Department, error) {
	var dt *data.Department
	var err error
	session := d.conn.Session(ctx)
	err = session.Where("id=?", id).Find(&dt).Error
	return dt, err
}

func NewDepartmentDao() *DepartmentDao {
	return &DepartmentDao{
		conn: gorms.New(),
	}
}
