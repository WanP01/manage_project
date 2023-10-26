package domain

import (
	"context"
	"project-grpc/user/login"
	"project-project/internal/rpc"
)

type UserRpcDomain struct {
	lc login.LoginServiceClient
}

func NewUserRpcDomain() *UserRpcDomain {
	return &UserRpcDomain{
		lc: rpc.UserGrpcClient,
	}
}

func (d *UserRpcDomain) MemberList(ctx context.Context, mIdList []int64) (*login.MemberMessageList, map[int64]*login.MemberMessage, error) {
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	messageList, err := d.lc.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	return messageList, mMap, err
}

func (d *UserRpcDomain) MemberInfo(ctx context.Context, memberCode int64) (*login.MemberMessage, error) {
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	memberMessage, err := d.lc.FindMemberById(ctx, &login.UserMessage{MemId: memberCode})
	return memberMessage, err
}
