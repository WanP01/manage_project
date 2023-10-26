package department_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"project-common/encrypts"
	"project-common/errs"
	"project-grpc/department"
	"project-project/domain"
	"project-project/internal/dao"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
)

type DepartmentService struct {
	department.UnimplementedDepartmentServiceServer
	cache            repo.Cache
	transaction      tran.Transaction
	departmentDomain *domain.DepartmentDomain
}

func New() *DepartmentService {
	return &DepartmentService{
		cache:            dao.Rc,
		transaction:      dao.NewTransactionDao(),
		departmentDomain: domain.NewDepartmentDomain(),
	}
}

func (d *DepartmentService) List(ctx context.Context, msg *department.DepartmentReqMessage) (*department.ListDepartmentMessage, error) {
	organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	var parentDepartmentCode int64
	if msg.ParentDepartmentCode != "" {
		parentDepartmentCode = encrypts.DecryptNoErr(msg.ParentDepartmentCode)
	}
	dps, total, err := d.departmentDomain.List(
		ctx,
		organizationCode,
		parentDepartmentCode,
		msg.Page,
		msg.PageSize)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var list []*department.DepartmentMessage
	copier.Copy(&list, dps)
	return &department.ListDepartmentMessage{List: list, Total: total}, nil
}

func (d *DepartmentService) Save(ctx context.Context, msg *department.DepartmentReqMessage) (*department.DepartmentMessage, error) {
	organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	var departmentCode int64
	if msg.DepartmentCode != "" {
		departmentCode = encrypts.DecryptNoErr(msg.DepartmentCode)
	}
	var parentDepartmentCode int64
	if msg.ParentDepartmentCode != "" {
		parentDepartmentCode = encrypts.DecryptNoErr(msg.ParentDepartmentCode)
	}
	dp, err := d.departmentDomain.Save(
		ctx,
		organizationCode,
		departmentCode,
		parentDepartmentCode,
		msg.Name)
	if err != nil {
		return &department.DepartmentMessage{}, errs.GrpcError(err)
	}
	var res = &department.DepartmentMessage{}
	copier.Copy(res, dp)
	return res, nil
}

func (d *DepartmentService) Read(ctx context.Context, msg *department.DepartmentReqMessage) (*department.DepartmentMessage, error) {
	//organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	departmentCode := encrypts.DecryptNoErr(msg.DepartmentCode)
	dp, err := d.departmentDomain.FindDepartmentById(ctx, departmentCode)
	if err != nil {
		return &department.DepartmentMessage{}, errs.GrpcError(err)
	}
	var res = &department.DepartmentMessage{}
	copier.Copy(res, dp.ToDisplay())
	return res, nil
}
