package dao

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type ProjectLogDao struct {
	conn *gorms.GormConn
}

func (p *ProjectLogDao) FindLogByMemberCode(ctx context.Context, memberId int64, page int64, size int64) ([]*data.ProjectLog, int64, error) {
	var list []*data.ProjectLog
	var total int64
	var err error
	session := p.conn.Session(ctx)
	offset := (page - 1) * size
	err = session.Model(&data.ProjectLog{}).
		Where("member_code=?", memberId).
		Limit(int(size)).
		Offset(int(offset)).Order("create_time desc").Find(&list).Error
	err = session.Model(&data.ProjectLog{}).
		Where("member_code=?", memberId).Count(&total).Error
	return list, total, err
}

func (p *ProjectLogDao) SaveProjectLog(pl *data.ProjectLog) {
	session := p.conn.Session(context.Background())
	session.Save(&pl)
}

// FindLogByTaskCode 查询全部
func (p *ProjectLogDao) FindLogByTaskCode(ctx context.Context, taskCode int64, comment int) ([]*data.ProjectLog, int64, error) {
	var list []*data.ProjectLog
	var total int64
	var err error
	session := p.conn.Session(ctx)
	model := session.Model(&data.ProjectLog{})
	if comment == 1 {
		model.Where("source_code=? and is_comment=?", taskCode, comment).Find(&list)
		model.Where("source_code=? and is_comment=?", taskCode, comment).Count(&total)
	} else {
		model.Where("source_code=?", taskCode).Find(&list)
		model.Where("source_code=?", taskCode).Count(&total)
	}
	return list, total, err
}

// FindLogByTaskCodePage 查询分页
func (p *ProjectLogDao) FindLogByTaskCodePage(ctx context.Context, taskCode int64, comment int, page int, pageSize int) ([]*data.ProjectLog, int64, error) {
	var list []*data.ProjectLog
	var total int64
	var err error
	session := p.conn.Session(ctx)
	model := session.Model(&data.ProjectLog{})
	offset := (page - 1) * pageSize
	if comment == 1 {
		model.Where("source_code=? and is_comment=?", taskCode, comment).Limit(pageSize).Offset(offset).Find(&list)
		model.Where("source_code=? and is_comment=?", taskCode, comment).Count(&total)
	} else {
		model.Where("source_code=?", taskCode).Limit(pageSize).Offset(offset).Find(&list)
		model.Where("source_code=?", taskCode).Count(&total)
	}
	return list, total, err
}

func NewProjectLogDao() *ProjectLogDao {
	return &ProjectLogDao{
		conn: gorms.New(),
	}
}
