package dao

import (
	"context"
	"project-project/internal/data/task"
	"project-project/internal/database/gorms"
)

type TaskStagesTemplateDao struct {
	conn *gorms.GormConn
}

func NewTaskStagesTemplateDao() *TaskStagesTemplateDao {
	return &TaskStagesTemplateDao{
		conn: gorms.New(),
	}
}

func (t *TaskStagesTemplateDao) FindInProTemIds(ctx context.Context, ids []int) ([]task.MsTaskStagesTemplate, error) {
	var tsts []task.MsTaskStagesTemplate
	err := t.conn.Session(ctx).Model(&task.MsTaskStagesTemplate{}).
		Where("project_template_code in ?", ids).
		Find(&tsts).Error
	return tsts, err
}
