package domain

import (
	"context"
	"project-grpc/auth"
	"project-user/internal/rpc"
)

type AuthRpcDomain struct {
	lc auth.AuthServiceClient
}

func NewAuthRpcDomain() *AuthRpcDomain {
	return &AuthRpcDomain{
		lc: rpc.AuthGrpcClient,
	}
}

func (ad *AuthRpcDomain) AuthList(ctx context.Context, msg *auth.AuthReqMessage) (*auth.ListAuthMessage, error) {
	authMessage, err := ad.lc.AuthList(ctx, msg)
	return authMessage, err
}

func (ad *AuthRpcDomain) Apply(ctx context.Context, msg *auth.AuthReqMessage) (*auth.ApplyResponse, error) {
	applyMessage, err := ad.lc.Apply(ctx, msg)
	return applyMessage, err
}

func (ad *AuthRpcDomain) AuthSave(ctx context.Context, msg *auth.AuthSaveReq) (*auth.ProjectAuth, error) {
	AuthMessage, err := ad.lc.AuthSave(ctx, msg)
	return AuthMessage, err
}
