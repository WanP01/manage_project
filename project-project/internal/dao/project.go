package dao

import (
	"context"
	"fmt"
	"project-project/internal/data/pro"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

func (p *ProjectDao) DeleteProjectCollect(ctx context.Context, memId int64, projectCode int64) error {
	return p.conn.Session(ctx).Where("member_code=? and project_code=?", memId, projectCode).Delete(&pro.ProjectCollection{}).Error
}

func (p *ProjectDao) SaveProjectCollect(ctx context.Context, pc *pro.ProjectCollection) error {
	return p.conn.Session(ctx).Save(&pc).Error
}

type ProjectDao struct {
	conn *gorms.GormConn
}

func (p *ProjectDao) UpdateProject(ctx context.Context, pj *pro.Project) error {
	return p.conn.Session(ctx).Updates(&pj).Error
}

func (p *ProjectDao) UpdateDeletedProject(ctx context.Context, projectCode int64, deleted bool) error {
	session := p.conn.Session(ctx)
	var err error
	if deleted {
		err = session.Model(&pro.Project{}).Where("id=?", projectCode).Update("deleted", 1).Error
	} else {
		err = session.Model(&pro.Project{}).Where("id=?", projectCode).Update("deleted", 0).Error
	}
	return err
}

func (p *ProjectDao) FindProjectByPIDANDMemID(ctx context.Context, memId int64, pId int64) (*pro.ProjectAndMember, error) {
	var pms *pro.ProjectAndMember
	//err := p.conn.Session(ctx).Model(&pro.Project{}).
	//	Joins("JOIN ms_project_member on ms_project.id = ms_project_member.project_code AND ms_project_member.member_code=? AND ms_project_member.project_code=? ", memId, pId).
	//	First(&pms).Error

	//避免 a 和 b 的 id值重复，从而 b Id 覆盖 a
	sql := fmt.Sprintf("select a.*,b.project_code,b.member_code,b.join_time,b.is_owner,b.authorize  from ms_project a, ms_project_member b where a.id = b.project_code and b.member_code=? and b.project_code=?")
	raw := p.conn.Session(ctx).Raw(sql, memId, pId)
	err := raw.Scan(&pms).Error
	return pms, err
}

func (p *ProjectDao) FindCollectProjectByPIDANDMemID(ctx context.Context, memId int64, pId int64) (bool, error) {
	var collect int64
	err := p.conn.Session(ctx).Model(&pro.ProjectMember{}).
		Where("member_code=? and project_code=? ", memId, pId).
		Count(&collect).Error
	return collect > 0, err
}

func (p *ProjectDao) SaveProject(conn database.DbConn, ctx context.Context, pr *pro.Project) error {
	p.conn = conn.(*gorms.GormConn)
	err := p.conn.Tx(ctx).Save(&pr).Error
	return err
}

func (p *ProjectDao) SaveProjectMember(conn database.DbConn, ctx context.Context, pm *pro.ProjectMember) error {
	p.conn = conn.(*gorms.GormConn)
	err := p.conn.Tx(ctx).Save(&pm).Error
	return err
}

func (p *ProjectDao) FindProjectByMemID(ctx context.Context, condition string, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error) {
	var pms []*pro.ProjectAndMember
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	sql := fmt.Sprintf("select a.*,b.project_code,b.member_code,b.join_time,b.is_owner,b.authorize  from ms_project a, ms_project_member b where a.id=b.project_code and b.member_code=? %s order by sort limit ?,?", condition) //避免 a 和 b 的 id值重复，从而 b Id 覆盖 a
	raw := session.Raw(sql, memId, index, size)                                                                                                                                                                                  //例如：第六页的30条
	raw.Scan(&pms)
	//_ = session.Table("ms_project").Joins("JOIN ms_project_member on ms_project.id = ms_project_member.project_code and ms_project_member.member_code=?", memId).Limit(int(size)).Offset(int(index)).Order("sort").Scan(&pms).Error
	var total int64
	query := fmt.Sprintf("select Count(*) from ms_project a,ms_project_member b where a.id=b.project_code and b.member_code=? %s", condition)
	tx := session.Raw(query, memId)
	err := tx.Scan(&total).Error
	return pms, total, err
}

func (p *ProjectDao) FindCollectProjectByMemID(ctx context.Context, condition string, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error) {

	var pms []*pro.ProjectAndMember
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	sql := fmt.Sprintf("select * from ms_project where id in (select project_code from ms_project_collection where member_code=?) %s order by sort limit ?,?", condition)
	raw := session.Raw(sql, memId, index, size) //例如：第六页的30条
	err := raw.Scan(&pms).Error

	var total int64
	query := fmt.Sprintf("select Count(*) from ms_project where id in (select project_code from ms_project_collection where member_code=?) %s ", condition)
	tx := session.Raw(query, memId)
	err = tx.Scan(&total).Error
	return pms, total, err
}

func NewProjectDao() *ProjectDao {
	return &ProjectDao{
		conn: gorms.New(),
	}
}
