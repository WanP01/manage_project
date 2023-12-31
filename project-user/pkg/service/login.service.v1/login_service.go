package login_service_v1

import (
	context "context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/copier"
	"log"
	"project-common/encrypts"
	"project-common/errs"
	"project-common/jwts"
	"project-common/tms"
	"project-grpc/account"
	"project-grpc/auth"
	"project-grpc/user/login"
	"project-user/config"
	"project-user/internal/dao"
	"project-user/internal/data/member"
	"project-user/internal/data/organization"
	"project-user/internal/database"
	"project-user/internal/database/tran"
	"project-user/internal/domain"
	"project-user/internal/repo"
	"project-user/pkg/model"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type LoginService struct {
	login.UnimplementedLoginServiceServer
	cache            repo.Cache
	memberRepo       repo.MemberRepo
	organizationRepo repo.OrganizationRepo
	transaction      tran.Transaction
	projectRpcDomain *domain.ProjectRpcDomain
	accountRpcDomain *domain.AccountRpcDomain
	authRpcDomain    *domain.AuthRpcDomain
}

func New() *LoginService {
	return &LoginService{
		cache:            dao.Rc,
		memberRepo:       dao.NewMemberDao(),
		organizationRepo: dao.NewOrganizationDao(),
		transaction:      dao.NewTransactionDao(),
		projectRpcDomain: domain.NewProjectRpcDomain(),
		accountRpcDomain: domain.NewAccountRpcDomain(),
		authRpcDomain:    domain.NewAuthRpcDomain(),
	}
}

func (ls *LoginService) GetCaptcha(ctx context.Context, msg *login.CaptchaMessage) (*login.CaptchaResponse, error) {
	mobile := msg.Mobile
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
		err := ls.cache.Put(context.Background(), model.RegisterRedisKey+mobile, code, 5*time.Minute)
		if err != nil {
			log.Printf("验证码存入redis错误，cause by %v :", err)
		} else {
			log.Printf("短信发送成功，将手机号存入redis成功,REGISTER_%v:%v\n", mobile, code)
		}
	}()

	return &login.CaptchaResponse{Code: code}, nil
}

func (ls *LoginService) Register(ctx context.Context, msg *login.RegisterMessage) (*login.RegisterResponse, error) {
	//c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	//c := context.Background()
	// 1. 可以校验参数（也可以在调用GRPC 之前校验）
	// 2. 校验验证码
	res, err := ls.cache.Get(ctx, model.RegisterRedisKey+msg.Mobile)
	if err == redis.Nil { // key不存在
		zap.L().Error("Register redis get error", zap.Error(err))
		return nil, errs.GrpcError(model.CaptchaNotExist)
	}
	if err != nil { // redis获取数据出错
		zap.L().Error("Register redis get error", zap.Error(err))
		return nil, errs.GrpcError(model.RedisError)
	}
	if res != msg.Captcha { // 验证码不匹配
		return nil, errs.GrpcError(model.CaptchaError)
	}
	// 3. 校验业务逻辑（邮箱是否被注册/账号是否被注册/手机号是否被注册）
	//邮箱校验
	exit, err := ls.memberRepo.GetMemberByEmail(ctx, msg.Email)
	if err != nil {
		return nil, errs.GrpcError(model.DBError)
	}
	if exit {
		return nil, errs.GrpcError(model.EmailExist)
	}
	//账号校验
	exit, err = ls.memberRepo.GetMemBerByAccount(ctx, msg.Name)
	if err != nil {
		return nil, errs.GrpcError(model.DBError)
	}
	if exit {
		return nil, errs.GrpcError(model.AccountExist)
	}
	//手机号校验
	exit, err = ls.memberRepo.GetMemberByMobile(ctx, msg.Mobile)
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
			pwd := encrypts.Md5(msg.Password)
			mem := &member.Member{
				Account:       msg.Name,
				Password:      pwd,
				Name:          msg.Name,
				Mobile:        msg.Mobile,
				Email:         msg.Email,
				CreateTime:    time.Now().UnixMilli(),
				LastLoginTime: time.Now().UnixMilli(),
				Status:        model.Normal,
			}
			err = ls.memberRepo.SaveMember(conn, ctx, mem)
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
			err = ls.organizationRepo.SaveOrganization(conn, ctx, org)
			if err != nil {
				zap.L().Error("Register db SaveOrganization error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			//生成一个账户，账户的授权角色 是成员，新生成一个角色（如果没有），同时将此角色的授权node 生成
			// 1. 查询是否有对应organization_code 的 auth

			listAuthMessage, err := ls.authRpcDomain.AuthList(ctx, &auth.AuthReqMessage{OrganizationCode: encrypts.EncryptNoErr(org.Id)})
			if err != nil {
				zap.L().Error("Register db authRpcDomain.AuthList organizationCode error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}

			var DefaultAuth *auth.ProjectAuth

			//2. 不存在就新生成 auth 和对应 auth_node
			if listAuthMessage.List == nil {
				//2.1 生成角色
				// 查找默认的auth（无 organization_code）
				authMessage, err := ls.authRpcDomain.AuthList(ctx, &auth.AuthReqMessage{})
				if err != nil {
					zap.L().Error("Register db authRpcDomain.AuthList  No organizationCode error", zap.Error(err))
					return errs.GrpcError(model.DBError)
				}

				for _, v := range authMessage.List {
					authSaveReq := &auth.AuthSaveReq{}
					// 保存自己的 auth（加入organization_code & createBy）
					copier.Copy(&authSaveReq, v)
					authSaveReq.CreateBy = mem.Id
					authSaveReq.OrganizationCode = encrypts.EncryptNoErr(org.Id)

					newListAuthMessage, err := ls.authRpcDomain.AuthSave(ctx, authSaveReq)
					if err != nil {
						zap.L().Error("Register db authRpcDomain.AuthSave error", zap.Error(err))
						return errs.GrpcError(model.DBError)
					}

					// 保存默认的Auth（管理员权限）
					if newListAuthMessage.IsDefault == 1 {
						DefaultAuth = newListAuthMessage
					}

					//2.2 生成node_auth
					//获取默认auth（NoOrg）的check list表
					ApplyList, err := ls.authRpcDomain.Apply(ctx, &auth.AuthReqMessage{AuthId: v.Id, Action: "getnode"})
					if err != nil {
						zap.L().Error("Register db authRpcDomain.Apply \"getnode\" error", zap.Error(err))
						return errs.GrpcError(model.DBError)
					}

					// 插入自己的 auth_node表
					_, err = ls.authRpcDomain.Apply(ctx, &auth.AuthReqMessage{AuthId: newListAuthMessage.Id, Action: "save", Nodes: ApplyList.CheckedList})
					if err != nil {
						zap.L().Error("Register db authRpcDomain.Apply \"save\" error", zap.Error(err))
						return errs.GrpcError(model.DBError)
					}
				}
			}

			// 4.  生成账户（用户，角色相关联）
			accountReq := &account.AccountSaveReq{}
			copier.Copy(&accountReq, mem)
			accountReq.MemberCode = encrypts.EncryptNoErr(mem.Id)
			accountReq.OrganizationCode = encrypts.EncryptNoErr(org.Id)
			accountReq.Authorize = encrypts.EncryptNoErr(DefaultAuth.Id)
			accountReq.IsOwner = 1
			_, err = ls.accountRpcDomain.AccountSave(ctx, accountReq)
			if err != nil {
				zap.L().Error("Register db AccountSave error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			return nil
		})
	// 5. 返回结果
	return &login.RegisterResponse{}, err
}

func (ls *LoginService) Login(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	//c := context.Background()
	//1.先寻找登录的提交信息（username 和 password）是否存在
	pwd := encrypts.Md5(msg.Password)
	meminfo, err := ls.memberRepo.FindMember(ctx, msg.Account, pwd)
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
	memMsg.Code, _ = encrypts.EncryptInt64(memMsg.Id, model.AESKey)
	memMsg.LastLoginTime = tms.FormatByMill(meminfo.LastLoginTime)
	memMsg.CreateTime = tms.FormatByMill(meminfo.CreateTime)
	//查询对应成员的个人组织（organization）
	orgs, err := ls.organizationRepo.FindOrganizationByMemID(ctx, memMsg.Id)
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
	for _, v := range orgMsg {
		v.Code, _ = encrypts.EncryptInt64(v.Id, model.AESKey)
		v.OwnerCode = memMsg.Code
		v.CreateTime = tms.FormatByMill(organization.ToMap(orgs)[v.Id].CreateTime)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	//3. 用jwt生成token
	memIdStr := strconv.FormatInt(memMsg.Id, 10)
	exp := time.Duration(config.AppConf.Jc.AccessExp*3600*24) * time.Second
	rExp := time.Duration(config.AppConf.Jc.RefreshExp*3600*24) * time.Second
	ip := msg.Ip
	token, err := jwts.CreateToken(memIdStr, exp, config.AppConf.Jc.AccessSecret, rExp, config.AppConf.Jc.RefreshSecret, ip)
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
	//保存缓存
	go func() {
		memberJson, _ := json.Marshal(meminfo) //不建议直接用grpc 的 memberMessage struct ， json 会忽视 omitempty
		ls.cache.Put(context.Background(), model.MemberRedisKey+"::"+memIdStr, string(memberJson), exp)
		orgJson, _ := json.Marshal(orgs) //不建议直接用grpc 的 memberMessage struct ， json 会忽视 omitempty
		ls.cache.Put(context.Background(), model.MemberOrganizationRedisKey+"::"+memIdStr, string(orgJson), exp)
	}()

	// 回复grpc响应
	return &login.LoginResponse{
		Member:           memMsg,
		OrganizationList: orgMsg,
		TokenList:        tokenList,
	}, nil

}

func (ls *LoginService) TokenVerify(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	//c := context.Background()
	// 获取token并处理格式
	token := msg.Token
	ip := msg.Ip
	if strings.Contains(token, "bearer") {
		token = strings.ReplaceAll(token, "bearer ", "")
	}
	// 解析token，验证通过则取出存在内部的 memID
	memIDstr, err := jwts.ParseToken(token, config.AppConf.Jc.AccessSecret, ip)
	if err != nil {
		zap.L().Error("Login  TokenVerify error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	//memId, err := strconv.ParseInt(memIDstr, 10, 64)
	//if err != nil {
	//	zap.L().Error("TokenVerify ParseInt err", zap.Error(err))
	//	return nil, errs.GrpcError(model.NoLogin)
	//}

	// 通过memID在数据库搜索 对应用户信息（优化前）
	// 优化点 登录之后 应该把用户信息缓存起来（优化后）
	memberJson, err := ls.cache.Get(ctx, model.MemberRedisKey+"::"+memIDstr)
	if err != nil {
		zap.L().Error("TokenVerify redis Get Member error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	if memberJson == "" {
		zap.L().Error("Login TokenVerify cache already expire")
		return nil, errs.GrpcError(model.NoLogin)
	}
	meminfo := &member.Member{}
	json.Unmarshal([]byte(memberJson), meminfo)

	memMsg := &login.MemberMessage{}
	copier.Copy(&memMsg, meminfo)

	//meminfo, err := ls.memberRepo.FindMemberByID(ctx, memId)
	//if err != nil {
	//	zap.L().Error("Login db FindMemByID error", zap.Error(err))
	//	return nil, errs.GrpcError(model.DBError)
	//}
	//if meminfo == nil {
	//	zap.L().Error("TokenVerify member is nil")
	//	return nil, errs.GrpcError(model.NoLogin)
	//}
	//// 返回mem信息即可（login 有organization 和 member和 tokenlist）
	//memMsg := &login.MemberMessage{}
	//err = copier.Copy(memMsg, meminfo)
	//if err != nil {
	//	zap.L().Error("memMsg copy error", zap.Error(err))
	//	return nil, errs.GrpcError(model.SyntaxError)
	//}
	memMsg.Code, _ = encrypts.EncryptInt64(memMsg.Id, model.AESKey)
	memMsg.LastLoginTime = tms.FormatByMill(meminfo.LastLoginTime)
	memMsg.CreateTime = tms.FormatByMill(meminfo.CreateTime)

	orgsJson, err := ls.cache.Get(ctx, model.MemberOrganizationRedisKey+"::"+memIDstr)
	if err != nil {
		zap.L().Error("TokenVerify redis Get MemberOrganization error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	if orgsJson == "" {
		zap.L().Error("Login TokenVerify cache already expire")
		return nil, errs.GrpcError(model.NoLogin)
	}
	var orgs []*organization.Organization
	json.Unmarshal([]byte(orgsJson), &orgs)

	//orgs, err := ls.organizationRepo.FindOrganizationByMemID(ctx, memMsg.Id)
	//if err != nil {
	//	zap.L().Error("TokenVerify db FindMember error", zap.Error(err))
	//	return nil, errs.GrpcError(model.DBError)
	//}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	return &login.LoginResponse{Member: memMsg}, nil
}

func (ls *LoginService) MyOrgList(ctx context.Context, msg *login.UserMessage) (*login.OrgListResponse, error) {
	//c := context.Background()
	//获取memId
	memId := msg.MemId
	//数据库查询organization
	orgs, err := ls.organizationRepo.FindOrganizationByMemID(ctx, memId)
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
	memMsgCode, _ := encrypts.EncryptInt64(memId, model.AESKey)
	for _, v := range orgMsg {
		v.Code, _ = encrypts.EncryptInt64(v.Id, model.AESKey)
		v.OwnerCode = memMsgCode
		v.CreateTime = tms.FormatByMill(organization.ToMap(orgs)[v.Id].CreateTime)
	}
	return &login.OrgListResponse{OrganizationList: orgMsg}, nil
}

func (ls *LoginService) FindMemberById(ctx context.Context, msg *login.UserMessage) (*login.MemberMessage, error) {
	memId := msg.MemId
	meminfo, err := ls.memberRepo.FindMemberByID(ctx, memId)
	if err != nil {
		zap.L().Error("Login FindMemberById FindMemByID error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 返回mem信息即可（login 有organization 和 member和 tokenlist）
	memMsg := &login.MemberMessage{}
	err = copier.Copy(memMsg, meminfo)
	if err != nil {
		zap.L().Error("memMsg copy error", zap.Error(err))
		return nil, errs.GrpcError(model.SyntaxError)
	}
	memMsg.Code, _ = encrypts.EncryptInt64(memMsg.Id, model.AESKey)
	memMsg.LastLoginTime = tms.FormatByMill(meminfo.LastLoginTime)
	memMsg.CreateTime = tms.FormatByMill(meminfo.CreateTime)
	orgs, err := ls.organizationRepo.FindOrganizationByMemID(ctx, memMsg.Id)
	if err != nil {
		zap.L().Error("TokenVerify db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	return memMsg, nil
}

func (ls *LoginService) FindMemInfoByIds(ctx context.Context, msg *login.UserMessage) (*login.MemberMessageList, error) {
	mIds := msg.MIds
	memberList, err := ls.memberRepo.FindMemberByIds(ctx, mIds)
	if err != nil {
		zap.L().Error("FindMemInfoByIds db memberRepo.FindMemberByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if memberList == nil || len(memberList) <= 0 {
		return &login.MemberMessageList{List: nil}, nil
	}
	mMap := make(map[int64]*member.Member)
	for _, v := range memberList {
		mMap[v.Id] = v
	}
	var memMsg []*login.MemberMessage
	copier.Copy(&memMsg, memberList)
	for _, v := range memMsg {
		v.Code, _ = encrypts.EncryptInt64(v.Id, model.AESKey)
		v.LastLoginTime = tms.FormatByMill(mMap[v.Id].LastLoginTime)
		v.CreateTime = tms.FormatByMill(mMap[v.Id].CreateTime)
	}
	return &login.MemberMessageList{List: memMsg}, nil
}
