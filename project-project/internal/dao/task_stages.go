package dao

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

type TaskStagesDao struct {
	conn *gorms.GormConn
}

func (t *TaskStagesDao) FindById(ctx context.Context, id int) (*data.TaskStages, error) {
	var ts *data.TaskStages
	err := t.conn.Session(ctx).Model(&data.TaskStages{}).Where("id=?", id).First(&ts).Error
	return ts, err
}

func (t *TaskStagesDao) FindTaskByProjectId(ctx context.Context, projectCode int64, page int64, size int64) ([]*data.TaskStages, int64, error) {
	var tst []*data.TaskStages
	session := t.conn.Session(ctx)
	err := session.Model(&data.TaskStages{}).
		Where("project_code=?", projectCode).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Order("sort asc").
		Find(&tst).Error

	var total int64
	err = session.Model(&data.TaskStages{}).
		Where("project_code=?", projectCode).
		Count(&total).Error
	return tst, total, err
}

func (t *TaskStagesDao) SaveTaskStages(ctx context.Context, conn database.DbConn, ts *data.TaskStages) error {
	t.conn = conn.(*gorms.GormConn)
	err := t.conn.Tx(ctx).Save(&ts).Error
	return err
}

func NewTaskStagesDao() *TaskStagesDao {
	return &TaskStagesDao{
		conn: gorms.New(),
	}
}
