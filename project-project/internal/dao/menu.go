package dao

import (
	"context"
	"project-project/internal/data/menu"
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

func (md *MenuDao) FindMenus(ctx context.Context) ([]*menu.ProjectMenu, error) {
	var menus []*menu.ProjectMenu
	err := md.conn.Session(ctx).Order("pid,sort asc,id asc").Find(&menus).Error
	return menus, err
}
