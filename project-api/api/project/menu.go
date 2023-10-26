package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/api/grpc"
	"project-api/pkg/model/menus"
	common "project-common"
	"project-common/errs"
	"project-grpc/menu"
)

type HandlerMenu struct {
}

func (m HandlerMenu) menuList(ctx *gin.Context) {
	result := &common.Result{}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background()
	res, err := grpc.MenuGrpcClient.MenuList(c, &menu.MenuReqMessage{})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var list []*menus.Menu
	copier.Copy(&list, res.List)
	if list == nil {
		list = []*menus.Menu{}
	}
	ctx.JSON(http.StatusOK, result.Success(list))
}

func NewMenu() *HandlerMenu {
	return &HandlerMenu{}
}
