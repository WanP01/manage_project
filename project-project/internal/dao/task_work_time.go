package dao

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type TaskWorkTimeDao struct {
	conn *gorms.GormConn
}

func (t *TaskWorkTimeDao) Save(ctx context.Context, twt *data.TaskWorkTime) error {
	session := t.conn.Session(ctx)
	err := session.Save(&twt).Error
	return err
}

func (t *TaskWorkTimeDao) FindWorkTimeList(ctx context.Context, taskCode int64) ([]*data.TaskWorkTime, error) {
	var list []*data.TaskWorkTime
	var err error
	session := t.conn.Session(ctx)
	err = session.Model(&data.TaskWorkTime{}).Where("task_code=?", taskCode).Find(&list).Error
	return list, err
}

func NewTaskWorkTimeDao() *TaskWorkTimeDao {
	return &TaskWorkTimeDao{
		conn: gorms.New(),
	}
}
