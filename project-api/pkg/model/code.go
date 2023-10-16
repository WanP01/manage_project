package model

import (
	"project-common/errs"
)

var (
	NoLegalMobile = errs.NewError(2001, "手机号不合法") //手机号格式不合法
)
