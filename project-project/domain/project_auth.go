package domain

import (
	"context"
	"go.uber.org/zap"
	"project-common/errs"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/repo"
	"project-project/pkg/model"
	"strconv"
)

type ProjectAuthDomain struct {
	projectAuthRepo       repo.ProjectAuthRepo
	userRpcDomain         *UserRpcDomain
	projectNodeDomain     *ProjectNodeDomain
	projectAuthNodeDomain *ProjectAuthNodeDomain
	memberAccountDomain   *MemberAccountDomain
}

func NewProjectAuthDomain() *ProjectAuthDomain {
	return &ProjectAuthDomain{
		projectAuthRepo:       dao.NewProjectAuthDao(),
		userRpcDomain:         NewUserRpcDomain(),
		projectNodeDomain:     NewProjectNodeDomain(),
		projectAuthNodeDomain: NewProjectAuthNodeDomain(),
		memberAccountDomain:   NewMemberAccountDomain(),
	}
}

func (d *ProjectAuthDomain) AuthList(ctx context.Context, orgCode int64) ([]*data.ProjectAuthDisplay, *errs.BError) {
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	list, err := d.projectAuthRepo.FindAuthList(ctx, orgCode)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, nil
}

func (d *ProjectAuthDomain) AuthListPage(ctx context.Context, orgCode int64, page int64, pageSize int64) ([]*data.ProjectAuthDisplay, int64, *errs.BError) {
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	list, total, err := d.projectAuthRepo.FindAuthListPage(ctx, orgCode, page, pageSize)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, 0, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, total, nil
}

func (d *ProjectAuthDomain) AllNodeAndAuth(ctx context.Context, authId int64) ([]*data.ProjectNodeAuthTree, []string, *errs.BError) {
	nodeList, err := d.projectNodeDomain.NodeList(ctx)
	if err != nil {
		return nil, nil, err
	}
	checkedList, err := d.projectAuthNodeDomain.AuthNodeList(ctx, authId)
	if err != nil {
		return nil, nil, err
	}
	list := data.ToAuthNodeTreeList(nodeList, checkedList)
	return list, checkedList, nil
}

func (d *ProjectAuthDomain) Save(ctx context.Context, conn database.DbConn, authId int64, nodes []string) *errs.BError {
	err := d.projectAuthNodeDomain.Save(ctx, conn, authId, nodes)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProjectAuthDomain) AuthNodes(ctx context.Context, memberId int64) ([]string, *errs.BError) {
	account, err := d.memberAccountDomain.FindAccount(ctx, memberId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, model.ParamsError
	}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	authorize := account.Authorize
	authId, _ := strconv.ParseInt(authorize, 10, 64)
	authNodeList, dbErr := d.projectAuthNodeDomain.AuthNodeList(ctx, authId)
	if dbErr != nil {
		return nil, model.DBError
	}
	return authNodeList, nil
}
