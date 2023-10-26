package menu_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"project-common/errs"
	"project-grpc/menu"
	"project-project/domain"
	"project-project/internal/dao"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
)

type MenuService struct {
	menu.UnimplementedMenuServiceServer
	cache       repo.Cache
	transaction tran.Transaction
	menuDomain  *domain.MenuDomain
}

func New() *MenuService {
	return &MenuService{
		cache:       dao.Rc,
		transaction: dao.NewTransactionDao(),
		menuDomain:  domain.NewMenuDomain(),
	}
}

func (ms *MenuService) MenuList(ctx context.Context, msg *menu.MenuReqMessage) (*menu.MenuResponseMessage, error) {
	ml, err := ms.menuDomain.MenuList(ctx)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var rsp []*menu.MenuMessage
	copier.Copy(&rsp, ml)
	return &menu.MenuResponseMessage{List: rsp}, nil
}
