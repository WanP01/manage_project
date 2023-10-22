package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

type TaskDao struct {
	conn *gorms.GormConn
}

func (t *TaskDao) FindTaskMemberPage(ctx context.Context, taskCode int64, page int64, size int64) ([]*data.TaskMember, int64, error) {
	var tmList []*data.TaskMember
	var total int64
	session := t.conn.Session(ctx)
	offset := (page - 1) * size
	err := session.Model(&data.TaskMember{}).Where("task_code=?", taskCode).Limit(int(size)).Offset(int(offset)).Find(&tmList).Error
	err = session.Model(&data.TaskMember{}).Where("task_code=?", taskCode).Count(&total).Error
	return tmList, total, err
}

func (t *TaskDao) FindTaskByAssignTo(ctx context.Context, memberId int64, done int, page int64, size int64) ([]*data.Task, int64, error) {
	var tsList []*data.Task
	var total int64
	session := t.conn.Session(ctx)
	offset := (page - 1) * size
	err := session.Model(&data.Task{}).Where("assign_to=? and deleted=0 and done=?", memberId, done).Limit(int(size)).Offset(int(offset)).Find(&tsList).Error
	err = session.Model(&data.Task{}).Where("assign_to=? and deleted=0 and done=?", memberId, done).Count(&total).Error
	return tsList, total, err
}

func (t *TaskDao) FindTaskByMemberCode(ctx context.Context, memberId int64, done int, page int64, size int64) ([]*data.Task, int64, error) {
	var tsList []*data.Task
	var total int64
	session := t.conn.Session(ctx)
	offset := (page - 1) * size
	sql := "select a.* from ms_task a,ms_task_member b where a.id=b.task_code and member_code=? and a.deleted=0 and a.done=? limit ?,?"
	raw := session.Model(&data.Task{}).Raw(sql, memberId, done, offset, size)
	err := raw.Scan(&tsList).Error
	if err != nil {
		return nil, 0, err
	}
	sqlCount := "select count(*) from ms_task a,ms_task_member b where a.id=b.task_code and member_code=? and a.deleted=0 and a.done=?"
	rawCount := session.Model(&data.Task{}).Raw(sqlCount, memberId, done)
	err = rawCount.Scan(&total).Error
	return tsList, total, err
}

func (t *TaskDao) FindTaskByCreateBy(ctx context.Context, memberId int64, done int, page int64, size int64) ([]*data.Task, int64, error) {
	var tsList []*data.Task
	var total int64
	session := t.conn.Session(ctx)
	offset := (page - 1) * size
	err := session.Model(&data.Task{}).Where("create_by=? and deleted=0 and done=?", memberId, done).Limit(int(size)).Offset(int(offset)).Find(&tsList).Error
	err = session.Model(&data.Task{}).Where("create_by=? and deleted=0 and done=?", memberId, done).Count(&total).Error
	return tsList, total, err
}

func (t *TaskDao) FindTaskByStageCodeLtSort(ctx context.Context, stageCode int, NextTaskSort int) (*data.Task, error) {
	var ts *data.Task
	err := t.conn.Session(ctx).Model(&data.Task{}).
		Where("stage_code=? and sort < ?", stageCode, NextTaskSort).
		Order("sort desc").Limit(1).
		First(&ts).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return ts, err
}

func (t *TaskDao) UpdateTaskSort(ctx context.Context, conn database.DbConn, ts *data.Task) error {
	t.conn = conn.(*gorms.GormConn)
	err := t.conn.Tx(ctx).Select("sort", "stage_code").Updates(&ts).Error
	return err
}

func (t *TaskDao) FindTaskById(ctx context.Context, taskCode int64) (*data.Task, error) {
	var ts *data.Task
	err := t.conn.Session(ctx).Model(&data.Task{}).Where("id=?", taskCode).First(&ts).Error
	return ts, err
}

func (t *TaskDao) SaveTask(ctx context.Context, conn database.DbConn, ts *data.Task) error {
	t.conn = conn.(*gorms.GormConn)
	err := t.conn.Tx(ctx).Save(&ts).Error
	return err
}

func (t *TaskDao) SaveTaskMember(ctx context.Context, conn database.DbConn, tm *data.TaskMember) error {
	t.conn = conn.(*gorms.GormConn)
	err := t.conn.Tx(ctx).Save(&tm).Error
	return err
}

func (t TaskDao) FindTaskSort(ctx context.Context, projectCode int64, stageCode int64) (*int, error) {
	var v *int
	session := t.conn.Session(ctx)
	//select * from
	err := session.Model(&data.Task{}).
		Where("project_code=? and stage_code=?", projectCode, stageCode).
		Select("max(sort)").Scan(&v).Error
	return v, err
}

func (t TaskDao) FindTaskMaxIdNum(ctx context.Context, ProjectCode int64) (*int, error) {
	session := t.conn.Session(ctx)
	var v *int
	err := session.Model(&data.Task{}).
		Where("project_code=?", ProjectCode).
		Select("max(id_num)").
		Scan(&v).Error //需要用Scan， 自定义字段或聚合函数的情况下，该查询目标可能为空，不能用Find or take （会报错）
	return v, err
}

func (t TaskDao) FindTaskMemberByTaskId(ctx context.Context, taskCode int64, memberCode int64) ([]*data.TaskMember, error) {
	var tm []*data.TaskMember
	err := t.conn.Session(ctx).Model(&data.TaskMember{}).
		Where("task_code=? and member_code=?", taskCode, memberCode).
		Find(&tm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return tm, err
}

func (t TaskDao) FindTaskByStageCode(ctx context.Context, stageCode int) ([]*data.Task, error) {
	var ts []*data.Task
	err := t.conn.Session(ctx).Model(&data.Task{}).
		Where("stage_code=? and deleted=0", stageCode).
		Order("sort asc").
		Find(&ts).Error
	return ts, err
}

func NewTaskDao() *TaskDao {
	return &TaskDao{
		conn: gorms.New(),
	}
}
