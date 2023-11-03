package domain

import (
	"context"
	"project-grpc/account"
	"project-user/internal/rpc"
)

type AccountRpcDomain struct {
	lc account.AccountServiceClient
}

func NewAccountRpcDomain() *AccountRpcDomain {
	return &AccountRpcDomain{
		lc: rpc.AccountGrpcClient,
	}
}

func (ad *AccountRpcDomain) AccountSave(ctx context.Context, msg *account.AccountSaveReq) (*account.AccountResponse, error) {
	accountMessage, err := ad.lc.AccountSave(ctx, msg)
	return accountMessage, err
}
