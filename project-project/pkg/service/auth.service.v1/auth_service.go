package auth_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"project-common/encrypts"
	"project-common/errs"
	"project-grpc/auth"
	"project-project/domain"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
	"time"
)

type AuthService struct {
	auth.UnimplementedAuthServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	projectAuthDomain *domain.ProjectAuthDomain
}

func New() *AuthService {
	return &AuthService{
		cache:             dao.Rc,
		transaction:       dao.NewTransactionDao(),
		projectAuthDomain: domain.NewProjectAuthDomain(),
	}
}

func (as *AuthService) AuthList(ctx context.Context, msg *auth.AuthReqMessage) (*auth.ListAuthMessage, error) {
	if msg.OrganizationCode != "" {
		organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
		listPage, total, err := as.projectAuthDomain.AuthListPage(ctx, organizationCode, msg.Page, msg.PageSize)
		if err != nil {
			return nil, errs.GrpcError(err)
		}
		var prList []*auth.ProjectAuth
		copier.Copy(&prList, listPage)
		return &auth.ListAuthMessage{List: prList, Total: total}, nil
	} else {
		listPage, err := as.projectAuthDomain.AuthListNoOrg(ctx)
		if err != nil {
			return nil, errs.GrpcError(err)
		}
		var prList []*auth.ProjectAuth
		copier.Copy(&prList, listPage)
		return &auth.ListAuthMessage{List: prList}, nil
	}
}

func (as *AuthService) Apply(ctx context.Context, msg *auth.AuthReqMessage) (*auth.ApplyResponse, error) {
	if msg.Action == "getnode" {
		//获取列表
		list, checkedList, err := as.projectAuthDomain.AllNodeAndAuth(ctx, msg.AuthId)
		if err != nil {
			return nil, errs.GrpcError(err)
		}
		var prList []*auth.ProjectNodeMessage
		copier.Copy(&prList, list)
		return &auth.ApplyResponse{List: prList, CheckedList: checkedList}, nil
	}
	if msg.Action == "save" {
		//先删除 project_auth_node表 在新增  事务
		//保存
		nodes := msg.Nodes
		//先删在存 加事务
		authId := msg.AuthId
		err := as.transaction.Action(func(conn database.DbConn) error {
			err := as.projectAuthDomain.AuthNodeSave(ctx, conn, authId, nodes)
			return err
		})
		if err != nil {
			return nil, errs.GrpcError(err.(*errs.BError))
		}
	}
	return &auth.ApplyResponse{}, nil
}

func (as *AuthService) AuthNodesByMemberId(ctx context.Context, msg *auth.AuthReqMessage) (*auth.AuthNodesResponse, error) {
	list, err := as.projectAuthDomain.AuthNodes(ctx, msg.MemberId)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	return &auth.AuthNodesResponse{List: list}, nil
}

func (as *AuthService) AuthSave(ctx context.Context, msg *auth.AuthSaveReq) (*auth.ProjectAuth, error) {
	organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	pa := &data.ProjectAuth{
		OrganizationCode: organizationCode,
		Title:            msg.Title,
		CreateAt:         time.Now().UnixMilli(),
		Sort:             int(msg.Sort),
		Status:           int(msg.Status),
		Desc:             msg.Desc,
		CreateBy:         msg.CreateBy,
		IsDefault:        int(msg.IsDefault),
		Type:             msg.Type,
	}
	err := as.projectAuthDomain.AuthSave(ctx, pa)
	if err != nil {
		return nil, errs.GrpcError(err)
	}

	paT, err := as.projectAuthDomain.FindAuthByTitleAndOrgCode(ctx, msg.Title, organizationCode)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	prA := &auth.ProjectAuth{}
	copier.Copy(&prA, paT)
	return prA, nil

}
