package account_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"project-common/encrypts"
	"project-common/errs"
	"project-grpc/account"
	"project-project/domain"
	"project-project/internal/dao"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
)

type AccountService struct {
	account.UnimplementedAccountServiceServer
	cache               repo.Cache
	transaction         tran.Transaction
	memberAccountDomain *domain.MemberAccountDomain
	projectAuthDomain   *domain.ProjectAuthDomain
}

func New() *AccountService {
	return &AccountService{
		cache:               dao.Rc,
		transaction:         dao.NewTransactionDao(),
		memberAccountDomain: domain.NewMemberAccountDomain(),
		projectAuthDomain:   domain.NewProjectAuthDomain(),
	}
}

func (as *AccountService) Account(ctx context.Context, msg *account.AccountReqMessage) (*account.AccountResponse, error) {
	accountList, total, err := as.memberAccountDomain.AccountList(
		ctx,
		msg.OrganizationCode,
		msg.MemberId,
		msg.Page,
		msg.PageSize,
		msg.DepartmentCode,
		msg.SearchType)
	if err != nil {
		return &account.AccountResponse{}, errs.GrpcError(err)
	}
	authList, err := as.projectAuthDomain.AuthList(ctx, encrypts.DecryptNoErr(msg.OrganizationCode))
	if err != nil {
		return &account.AccountResponse{}, errs.GrpcError(err)
	}
	var maList []*account.MemberAccount
	copier.Copy(&maList, accountList)
	var prList []*account.ProjectAuth
	copier.Copy(&prList, authList)
	return &account.AccountResponse{
		AccountList: maList,
		AuthList:    prList,
		Total:       total,
	}, nil
}
