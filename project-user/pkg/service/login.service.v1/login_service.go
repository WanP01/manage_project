package login_service_v1

import (
	context "context"
	"log"
	common "project-common"
	"project-common/errs"
	"project-user/pkg/dao"
	"project-user/pkg/model"
	"project-user/pkg/repo"
	"time"

	"go.uber.org/zap"
)

type LoginService struct {
	UnimplementedLoginServiceServer
	cache repo.Cache
}

func New() *LoginService {
	return &LoginService{
		cache: dao.Rc,
	}
}

func (ls *LoginService) GetCaptcha(ctx context.Context, cm *CaptchaMessage) (*CaptchaResponse, error) {
	mobile := cm.Mobile
	// 校验参数（validate)
	if !common.VerifyMobile(mobile) {
		return nil, errs.GrpcError(model.NoLegalMobile)
	}
	//生成验证码（随机4位或6位数字）
	code := "12345"
	//调用短信平台（三方 go协程异步执行，接口快速响应）
	go func() {
		time.Sleep(2 * time.Second)
		zap.L().Info("短信平台调用成功,发送验证码 Info")
		// zap.L().Debug("短信平台调用成功,发送验证码 Debug")
		// zap.L().Warn("短信平台调用成功,发送验证码 Warn")
		//log.Printf("短信平台调用成功,发送验证码%v\n", code)
		//存储验证码redis当中，过期时间5min
		//注意点，后续存储的软件可能不一致，比如redis 或者其他nosql软件，所以需要用接口，降低代码耦合
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := ls.cache.Put(ctx, "REGISTER_"+mobile, code, 5*time.Minute)
		if err != nil {
			log.Printf("验证码存入redis错误，cause by %v :", err)
		} else {
			log.Printf("短信发送成功，将手机号存入redis成功,REGISTER_%v:%v\n", mobile, code)
		}
	}()

	return &CaptchaResponse{Code: code}, nil
}
