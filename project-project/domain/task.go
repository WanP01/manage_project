package domain

import (
	"context"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/repo"
	"project-project/pkg/model"
)

type TaskDomain struct {
	taskRepo repo.TaskRepo
}

func NewTaskDomain() *TaskDomain {
	return &TaskDomain{
		taskRepo: dao.NewTaskDao(),
	}
}

func (d *TaskDomain) FindProjectIdByTaskId(ctx context.Context, taskId int64) (int64, bool, *errs.BError) {
	task, err := d.taskRepo.FindTaskById(ctx, taskId)
	if err != nil {
		return 0, false, model.DBError
	}
	if task == nil {
		return 0, false, nil
	}
	return task.ProjectCode, true, nil
}
