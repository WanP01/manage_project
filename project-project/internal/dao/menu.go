package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type MenuDao struct {
	conn *gorms.GormConn
}

func NewMenuDao() *MenuDao {
	return &MenuDao{
		conn: gorms.New(),
	}
}

func (md *MenuDao) FindMenus(ctx context.Context) ([]*data.ProjectMenu, error) {
	var menus []*data.ProjectMenu
	err := md.conn.Session(ctx).Order("pid,sort asc,id asc").Find(&menus).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return menus, err
}
