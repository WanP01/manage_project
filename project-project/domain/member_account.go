package domain

import (
	"context"
	"fmt"
	"project-common/encrypts"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/repo"
	"project-project/pkg/model"
)

type MemberAccountDomain struct {
	memberAccountRepo repo.MemberAccountRepo
	userGrpcDomain    *UserRpcDomain
	departmentDomain  *DepartmentDomain
}

func (d MemberAccountDomain) AccountList(ctx context.Context, organizationCode string, memberId int64, page int64, pageSize int64, departmentCode string, searchType int32) ([]*data.MemberAccountDisplay, int64, *errs.BError) {
	condition := ""
	organizationCodeId := encrypts.DecryptNoErr(organizationCode)
	departmentCodeId := encrypts.DecryptNoErr(departmentCode)
	switch searchType {
	case 1:
		condition = "status = 1" //使用当中的account
	case 2:
		condition = "department_code = NULL" // 查询系统的account
	case 3:
		condition = "status = 0" // 查询禁用account
	case 4:
		condition = fmt.Sprintf("status = 1 and department_code = %d", departmentCodeId) // 查询正在使用的，当前部门下的account
	default:
		condition = "status = 1" //默认查询使用中的account
	}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()

	//查询账号列表
	list, total, err := d.memberAccountRepo.FindList(ctx, condition, organizationCodeId, departmentCodeId, page, pageSize)
	if err != nil {
		return nil, 0, model.DBError
	}
	var dList []*data.MemberAccountDisplay
	for _, v := range list {
		display := v.ToDisplay()
		// 查询用户信息
		memberInfo, _ := d.userGrpcDomain.MemberInfo(ctx, v.MemberCode)
		display.Avatar = memberInfo.Avatar
		display.Name = memberInfo.Name
		display.Mobile = memberInfo.Mobile
		display.Email = memberInfo.Email
		if v.DepartmentCode > 0 {
			department, err := d.departmentDomain.FindDepartmentById(ctx, v.DepartmentCode)
			if err != nil {
				return nil, 0, errs.ToBError(err)
			}
			display.Departments = department.Name
		}
		dList = append(dList, display)
	}
	return dList, total, nil
}

func (d *MemberAccountDomain) FindAccount(ctx context.Context, memberId int64) (*data.MemberAccount, *errs.BError) {
	account, err := d.memberAccountRepo.FindByMemberId(ctx, memberId)
	if err != nil {
		return nil, model.DBError
	}
	return account, nil
}

func (d *MemberAccountDomain) Save(ctx context.Context, memberAccount *data.MemberAccount) *errs.BError {
	err := d.memberAccountRepo.Save(ctx, memberAccount)
	if err != nil {
		return model.DBError
	}
	return nil
}

func NewMemberAccountDomain() *MemberAccountDomain {
	return &MemberAccountDomain{
		memberAccountRepo: dao.NewMemberAccountDao(),
		userGrpcDomain:    NewUserRpcDomain(),
		departmentDomain:  NewDepartmentDomain(),
	}
}
