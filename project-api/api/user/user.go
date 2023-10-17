package user

import (
	"context"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	"project-api/pkg/model/user"
	common "project-common"
	"project-common/errs"
	"project-grpc/user/login"
	"time"

	"github.com/gin-gonic/gin"
)

type HandlerUser struct {
}

func NewHandlerUser() *HandlerUser {
	return &HandlerUser{}
}

func (hu *HandlerUser) getCaptcha(ctx *gin.Context) {
	result := &common.Result{}
	mobile := ctx.PostForm("mobile")
	// 校验参数（validate)
	if !common.VerifyMobile(mobile) {
		ctx.JSON(http.StatusOK, result.Fail(common.BusinessCode(model.NoLegalMobile.Code), model.NoLegalMobile.Msg))
		return
	}
	//调用User模块的Grpc（验证码服务）
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res, err := UserGrpcClient.GetCaptcha(c, &login.CaptchaMessage{Mobile: mobile})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	//包装获得的数据反馈
	ctx.JSON(http.StatusOK, result.Success(res.GetCode()))
}

func (hu *HandlerUser) register(ctx *gin.Context) {
	// 1.接收参数 参数模型
	result := &common.Result{}
	var req user.RegisterReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}
	// 2. 校验参数 判断参数是否合法
	if err := req.Verify(); err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	// 3.调用User grpc服务 获取响应
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	c := context.Background() // 调试用
	// copier 库实现反射复制
	msg := &login.RegisterMessage{}
	err = copier.Copy(msg, req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	////处理业务（逐行赋值）
	//msg := &RegisterMessage{
	//	Name:     req.Name,
	//	Email:    req.Email,
	//	Mobile:   req.Mobile,
	//	Password: req.Password,
	//	Captcha:  req.Captcha,
	//}
	_, err = UserGrpcClient.Register(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	// 4. 返回响应
	ctx.JSON(http.StatusOK, result.Success(""))
	return
}

func (hu *HandlerUser) login(ctx *gin.Context) {
	// 1.接收参数 参数模型
	result := &common.Result{}
	var req user.LoginReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}
	// 2.调用user grpc 完成登录
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//c := context.Background() // 调试用
	msg := &login.LoginMessage{}
	err = copier.Copy(msg, req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	loginResp, err := UserGrpcClient.Login(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	// 回复响应的用户数据
	rsp := &user.LoginRsp{}
	err = copier.Copy(rsp, loginResp)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(rsp))
}
