package dao

import (
	"context"
	"fmt"
	"project-project/internal/data/pro"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

type ProjectDao struct {
	conn *gorms.GormConn
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

func NewProjectDao() *ProjectDao {
	return &ProjectDao{
		conn: gorms.New(),
	}
}

func (p *ProjectDao) FindProjectByMemID(ctx context.Context, condition string, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error) {
	var pms []*pro.ProjectAndMember
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	sql := fmt.Sprintf("select * from ms_project a, ms_project_member b where a.id=b.project_code and b.member_code=? %s order by sort limit ?,?", condition)
	raw := session.Raw(sql, memId, index, size) //例如：第六页的30条
	raw.Scan(&pms)
	//_ = session.Table("ms_project").Joins("JOIN ms_project_member on ms_project.id = ms_project_member.project_code and ms_project_member.member_code=?", memId).Limit(int(size)).Offset(int(index)).Order("sort").Scan(&pms).Error
	var total int64
	query := fmt.Sprintf("select Count(*) from ms_project a,ms_project_member b where a.id=b.project_code and b.member_code=? %s", condition)
	tx := session.Raw(query, memId)
	err := tx.Scan(&total).Error
	return pms, total, err
}

func (p *ProjectDao) FindCollectProjectByMemID(ctx context.Context, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error) {
	var pms []*pro.ProjectAndMember
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	sql := fmt.Sprintf("select * from ms_project where id in (select project_code from ms_project_collection where member_code=?) order by sort limit ?,?")
	raw := session.Raw(sql, memId, index, size) //例如：第六页的30条
	err := raw.Scan(&pms).Error
	var total int64
	query := fmt.Sprintf("member_code=?")
	err = session.Model(&pro.ProjectCollection{}).Where(query, memId).Count(&total).Error
	return pms, total, err

}
