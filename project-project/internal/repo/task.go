package repo

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database"
)

// TaskStagesTemplateRepo task_stage 《=》 template
type TaskStagesTemplateRepo interface {
	FindInProTemIds(ctx context.Context, ids []int) ([]data.MsTaskStagesTemplate, error)
	FindByProjectTemplateId(ctx context.Context, id int) ([]*data.MsTaskStagesTemplate, error)
}

// TaskStagesRepo task_stage 《=》 Project
type TaskStagesRepo interface {
	SaveTaskStages(ctx context.Context, conn database.DbConn, ts *data.TaskStages) error
	FindTaskByProjectId(ctx context.Context, projectCode int64, page int64, size int64) ([]*data.TaskStages, int64, error)
	FindById(ctx context.Context, id int) (*data.TaskStages, error)
}

// TaskRepo task 《=》 task_stage
type TaskRepo interface {
	FindTaskByStageCode(ctx context.Context, stageCode int) ([]*data.Task, error)
	FindTaskMemberByTaskId(ctx context.Context, taskCode int64, memberCode int64) ([]*data.TaskMember, error)
	FindTaskMaxIdNum(ctx context.Context, ProjectCode int64) (*int, error)
	FindTaskSort(ctx context.Context, projectCode int64, stageCode int64) (*int, error)
	SaveTask(ctx context.Context, conn database.DbConn, ts *data.Task) error
	SaveTaskMember(ctx context.Context, conn database.DbConn, tm *data.TaskMember) error
	FindTaskById(ctx context.Context, taskCode int64) (*data.Task, error)
	UpdateTaskSort(ctx context.Context, conn database.DbConn, ts *data.Task) error
	FindTaskByStageCodeLtSort(ctx context.Context, stageCode int, NextTaskSort int) (*data.Task, error)
	FindTaskByAssignTo(ctx context.Context, memberId int64, done int, page int64, size int64) ([]*data.Task, int64, error)
	FindTaskByMemberCode(ctx context.Context, memberId int64, done int, page int64, size int64) ([]*data.Task, int64, error)
	FindTaskByCreateBy(ctx context.Context, memberId int64, done int, page int64, size int64) ([]*data.Task, int64, error)
	FindTaskMemberPage(ctx context.Context, taskCode int64, page int64, size int64) ([]*data.TaskMember, int64, error)
}
