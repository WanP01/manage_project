package user

import (
	"errors"
	common "project-common"
)

// RegisterReq 注册信息
type RegisterReq struct {
	Email     string `json:"email" form:"email"`
	Name      string `json:"name" form:"name"`
	Password  string `json:"password" form:"password"`
	Password2 string `json:"password2" form:"password2"`
	Mobile    string `json:"mobile" form:"mobile"`
	Captcha   string `json:"captcha" form:"captcha"`
}

// VerifyPassword 验证密码一致性
func (r *RegisterReq) VerifyPassword() bool {
	return r.Password == r.Password2
}

// Verify 校验注册信息格式合法
func (r *RegisterReq) Verify() error {
	if !common.VerifyEmailFormat(r.Email) {
		return errors.New("邮箱格式不对")
	}
	if !common.VerifyMobile(r.Mobile) {
		return errors.New("手机号格式不正确")
	}
	if !r.VerifyPassword() {
		return errors.New("两次密码输入不一致")
	}
	return nil
}

// LoginReq 登录请求
type LoginReq struct {
	Account  string `json:"account" form:"account"`
	Password string `json:"password" form:"password"`
}

// LoginRsp 登录响应
type LoginRsp struct {
	Member           Member             `json:"member"`
	TokenList        TokenList          `json:"tokenList"`
	OrganizationList []OrganizationList `json:"organizationList"`
}
type Member struct {
	//Id     int64  `json:"id"`// 避免暴露内部序号，需要加密，加密后为code字段
	Name   string `json:"name"`
	Mobile string `json:"mobile"`
	Status int    `json:"status"`
	Code   string `json:"code"`
}

type TokenList struct {
	AccessToken    string `json:"accessToken"`
	RefreshToken   string `json:"refreshToken"`
	TokenType      string `json:"tokenType"`
	AccessTokenExp int64  `json:"accessTokenExp"`
}

type OrganizationList struct {
	//Id          int64  `json:"id"` // 避免暴露内部序号，需要加密，加密后为code字段
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	MemberId    int64  `json:"memberId"`
	CreateTime  int64  `json:"createTime"`
	Personal    int32  `json:"personal"`
	Address     string `json:"address"`
	Province    int32  `json:"province"`
	City        int32  `json:"city"`
	Area        int32  `json:"area"`
	Code        string `json:"code"`
}
