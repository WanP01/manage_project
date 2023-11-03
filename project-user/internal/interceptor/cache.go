package interceptor

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"project-common/encrypts"
	"project-grpc/user/login"
	"project-user/internal/dao"
	"project-user/internal/repo"
	"time"
)

var (
	grpcReqRspMap = map[string]any{
		"/login.service.v1.LoginService/MyOrgList":      &login.OrgListResponse{},
		"/login.service.v1.LoginService/FindMemberById": &login.MemberMessage{},
	}
)

type CacheInterceptor struct {
	cache    repo.Cache
	cacheMap map[string]any // 不同请求（string）对应的返回结果类型（any）
}

type Options func(interceptor *CacheInterceptor)

func NewCacheInterceptor(opt ...Options) *CacheInterceptor {
	ci := &CacheInterceptor{
		cache:    dao.Rc,
		cacheMap: grpcReqRspMap, // map默认值
	}
	for _, o := range opt {
		o(ci)
	}
	return ci
}

// WithCacheMap 可修改默认值
func WithCacheMap(reqMap map[string]any) Options {
	return func(interceptor *CacheInterceptor) {
		interceptor.cacheMap = reqMap
	}
}

func (ci *CacheInterceptor) CacheInterceptorOpt() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		//重新New CacheInterceptor //主要是类似TaskList service 有不同 stage参数，但只会会走一个路径的缓存会导致缓存紊乱
		ci = NewCacheInterceptor()
		// 确认是否是定义的路径
		respType := ci.cacheMap[info.FullMethod]
		if respType == nil {
			return handler(ctx, req)
		}
		// 尝试拿取redis中的缓存
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		marshal, _ := json.Marshal(req)
		cacheKey := encrypts.Md5(string(marshal))
		key := info.FullMethod + "::" + cacheKey
		respJson, err := ci.cache.Get(c, key)
		//如果有就返回
		if respJson != "" {
			json.Unmarshal([]byte(respJson), &respType)
			zap.L().Info(info.FullMethod + " 走了缓存")
			return respType, nil
		}
		//没有就存入缓存
		res, err := handler(ctx, req)
		cacheResp, _ := json.Marshal(res)
		_ = ci.cache.Put(ctx, key, string(cacheResp), 5*time.Minute)
		zap.L().Info(info.FullMethod + " 放入缓存")
		return res, err
	})
}

func (ci *CacheInterceptor) CacheInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		//重新New CacheInterceptor //主要是类似TaskList service 有不同 stage参数，但只会会走一个路径的缓存会导致缓存紊乱
		ci = NewCacheInterceptor()
		// 确认是否是定义的路径
		respType := ci.cacheMap[info.FullMethod]
		if respType == nil {
			return handler(ctx, req)
		}
		// 尝试拿取redis中的缓存
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		marshal, _ := json.Marshal(req)
		cacheKey := encrypts.Md5(string(marshal))
		key := info.FullMethod + "::" + cacheKey
		respJson, err := ci.cache.Get(c, key)
		//如果有就返回
		if respJson != "" {
			json.Unmarshal([]byte(respJson), &respType)
			zap.L().Info(info.FullMethod + " 走了缓存")
			return respType, nil
		}
		//没有就存入缓存
		res, err := handler(ctx, req)
		cacheResp, _ := json.Marshal(res)
		_ = ci.cache.Put(ctx, key, string(cacheResp), 5*time.Minute)
		zap.L().Info(info.FullMethod + " 放入缓存")
		return res, err
	}
}
