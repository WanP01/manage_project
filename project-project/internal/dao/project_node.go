package dao

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type ProjectNodeDao struct {
	conn *gorms.GormConn
}

func (m *ProjectNodeDao) FindAll(ctx context.Context) ([]*data.ProjectNode, error) {
	var pms []*data.ProjectNode
	var err error
	session := m.conn.Session(ctx)
	err = session.Model(&data.ProjectNode{}).Find(&pms).Error
	return pms, err
}

func NewProjectNodeDao() *ProjectNodeDao {
	return &ProjectNodeDao{
		conn: gorms.New(),
	}
}
