package login_service_v1

import (
	context "context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/copier"
	"log"
	"project-common/encrypts"
	"project-common/errs"
	"project-common/jwts"
	"project-grpc/user/login"
	"project-user/config"
	"project-user/internal/dao"
	"project-user/internal/data/member"
	"project-user/internal/data/organization"
	"project-user/internal/database"
	"project-user/internal/database/tran"
	"project-user/internal/repo"
	"project-user/pkg/model"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type LoginService struct {
	login.UnimplementedLoginServiceServer
	cache            repo.Cache
	memberRepo       repo.MemberRepo
	organizationRepo repo.OrganizationRepo
	transaction      tran.Transaction
}

func New() *LoginService {
	return &LoginService{
		cache:            dao.Rc,
		memberRepo:       dao.NewMemberDao(),
		organizationRepo: dao.NewOrganizationDao(),
		transaction:      dao.NewTransactionDao(),
	}
}

func (ls *LoginService) GetCaptcha(ctx context.Context, cm *login.CaptchaMessage) (*login.CaptchaResponse, error) {
	mobile := cm.Mobile
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
		err := ls.cache.Put(ctx, model.RegisterRedisKey+mobile, code, 5*time.Minute)
		if err != nil {
			log.Printf("验证码存入redis错误，cause by %v :", err)
		} else {
			log.Printf("短信发送成功，将手机号存入redis成功,REGISTER_%v:%v\n", mobile, code)
		}
	}()

	return &login.CaptchaResponse{Code: code}, nil
}

func (ls *LoginService) Register(ctx context.Context, rm *login.RegisterMessage) (*login.RegisterResponse, error) {
	//c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	c := context.Background()
	// 1. 可以校验参数（也可以在调用GRPC 之前校验）
	// 2. 校验验证码
	res, err := ls.cache.Get(c, model.RegisterRedisKey+rm.Mobile)
	if err == redis.Nil { // key不存在
		zap.L().Error("Register redis get error", zap.Error(err))
		return nil, errs.GrpcError(model.CaptchaNotExist)
	}
	if err != nil { // redis获取数据出错
		zap.L().Error("Register redis get error", zap.Error(err))
		return nil, errs.GrpcError(model.RedisError)
	}
	if res != rm.Captcha { // 验证码不匹配
		return nil, errs.GrpcError(model.CaptchaError)
	}
	// 3. 校验业务逻辑（邮箱是否被注册/账号是否被注册/手机号是否被注册）
	//邮箱校验
	exit, err := ls.memberRepo.GetMemberByEmail(c, rm.Email)
	if err != nil {
		return nil, errs.GrpcError(model.DBError)
	}
	if exit {
		return nil, errs.GrpcError(model.EmailExist)
	}
	//账号校验
	exit, err = ls.memberRepo.GetMemBerByAccount(c, rm.Name)
	if err != nil {
		return nil, errs.GrpcError(model.DBError)
	}
	if exit {
		return nil, errs.GrpcError(model.AccountExist)
	}
	//手机号校验
	exit, err = ls.memberRepo.GetMemberByMobile(c, rm.Mobile)
	if err != nil {
		return nil, errs.GrpcError(model.DBError)
	}
	if exit {
		return nil, errs.GrpcError(model.MobileExist)
	}
	// 4. 执行业务 将数据存入 member 表，生成一个数据，并同步存入组织表 organization
	// 整体流程应当采用事务流程保持原子性和一致性
	err = ls.transaction.Action(
		func(conn database.DbConn) error {
			pwd := encrypts.Md5(rm.Password)
			mem := &member.Member{
				Account:       rm.Name,
				Password:      pwd,
				Name:          rm.Name,
				Mobile:        rm.Mobile,
				Email:         rm.Email,
				CreateTime:    time.Now().UnixMilli(),
				LastLoginTime: time.Now().UnixMilli(),
				Status:        model.Normal,
			}
			err = ls.memberRepo.SaveMember(conn, c, mem)
			if err != nil {
				zap.L().Error("Register db SaveMember error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			org := &organization.Organization{
				Name:       mem.Name + "个人组织",
				MemberId:   mem.Id,
				CreateTime: time.Now().UnixMilli(),
				Personal:   model.Personal,
				Avatar:     "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fc-ssl.dtstatic.com%2Fuploads%2Fblog%2F202103%2F31%2F20210331160001_9a852.thumb.1000_0.jpg&refer=http%3A%2F%2Fc-ssl.dtstatic.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1673017724&t=ced22fc74624e6940fd6a89a21d30cc5",
			}
			err = ls.organizationRepo.SaveOrganization(conn, c, org)
			if err != nil {
				zap.L().Error("Register db SaveOrganization error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			return nil
		})

	// 5. 返回结果
	return &login.RegisterResponse{}, err
}

func (ls *LoginService) Login(ctx context.Context, lm *login.LoginMessage) (*login.LoginResponse, error) {
	c := context.Background()
	//1.先寻找登录的提交信息（username 和 password）是否存在
	pwd := encrypts.Md5(lm.Password)
	meminfo, err := ls.memberRepo.FindMember(c, lm.Account, pwd)
	if err != nil {
		zap.L().Error("Login db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if meminfo == nil {
		return nil, errs.GrpcError(model.AccountAndPwdError)
	}
	memMsg := &login.MemberMessage{}
	err = copier.Copy(memMsg, meminfo)
	if err != nil {
		zap.L().Error("memMsg copy error", zap.Error(err))
		return nil, errs.GrpcError(model.SyntaxError)
	}
	//查询对应成员的个人组织（organization）
	orgs, err := ls.organizationRepo.FindOrganizationByMemID(c, memMsg.Id)
	if err != nil {
		zap.L().Error("Login db FindOrganizationByMemID error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var orgMsg []*login.OrganizationMessage
	err = copier.Copy(&orgMsg, orgs)
	if err != nil {
		zap.L().Error("memMsg copy error", zap.Error(err))
		return nil, errs.GrpcError(model.SyntaxError)
	}
	//3. 用jwt生成token
	memIdStr := strconv.FormatInt(memMsg.Id, 10)
	exp := time.Duration(config.AppConf.Jc.AccessExp*3600*24) * time.Second
	fmt.Println(config.AppConf.Jc.AccessExp)
	rExp := time.Duration(config.AppConf.Jc.RefreshExp*3600*24) * time.Second
	fmt.Println(config.AppConf.Jc.RefreshExp)
	token, err := jwts.CreateToken(memIdStr, exp, config.AppConf.Jc.AccessSecret, rExp, config.AppConf.Jc.RefreshSecret)
	if err != nil {
		zap.L().Error("Jwt Generate error", zap.Error(err))
		return nil, errs.GrpcError(model.SyntaxError)
	}
	tokenList := &login.TokenMessage{
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		AccessTokenExp: token.AccessExp,
		TokenType:      "bearer",
	}
	// 回复grpc响应
	return &login.LoginResponse{
		Member:           memMsg,
		OrganizationList: orgMsg,
		TokenList:        tokenList,
	}, nil

}
