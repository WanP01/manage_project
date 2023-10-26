package domain

import (
	"context"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/repo"
	"project-project/pkg/model"
	"time"
)

type DepartmentDomain struct {
	departmentRepo repo.DepartmentRepo
}

func (d *DepartmentDomain) FindDepartmentById(ctx context.Context, id int64) (*data.Department, *errs.BError) {
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	dp, err := d.departmentRepo.FindDepartmentById(ctx, id)
	if err != nil {
		return nil, model.DBError
	}
	return dp, nil
}

func (d *DepartmentDomain) List(ctx context.Context, organizationCode int64, parentDepartmentCode int64, page int64, size int64) ([]*data.DepartmentDisplay, int64, *errs.BError) {
	list, total, err := d.departmentRepo.ListDepartment(ctx, organizationCode, parentDepartmentCode, page, size)
	if err != nil {
		return nil, 0, model.DBError
	}
	var dList []*data.DepartmentDisplay
	for _, v := range list {
		dList = append(dList, v.ToDisplay())
	}
	return dList, total, nil
}

func (d *DepartmentDomain) Save(ctx context.Context,
	organizationCode int64,
	departmentCode int64,
	parentDepartmentCode int64,
	name string) (*data.DepartmentDisplay, *errs.BError) {

	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	dpm, err := d.departmentRepo.FindDepartment(ctx, organizationCode, parentDepartmentCode, name)
	if err != nil {
		return nil, model.DBError
	}
	if dpm == nil {
		dpm = &data.Department{
			Name:             name,
			OrganizationCode: organizationCode,
			CreateTime:       time.Now().UnixMilli(),
		}
		if parentDepartmentCode > 0 {
			dpm.Pcode = parentDepartmentCode
		}
		err := d.departmentRepo.Save(ctx, dpm)
		if err != nil {
			return nil, model.DBError
		}
		return dpm.ToDisplay(), nil
	}
	return dpm.ToDisplay(), nil
}

func NewDepartmentDomain() *DepartmentDomain {
	return &DepartmentDomain{
		departmentRepo: dao.NewDepartmentDao(),
	}
}
