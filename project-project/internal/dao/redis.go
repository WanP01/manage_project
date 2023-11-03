package dao

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var Rc *RedisCache

//改为在 配置读取阶段 建立连接 （config）
//func init() {
//	rdb := redis.NewClient(config.AppConf.InitRedisOptions())
//
//	Rc = &RedisCache{
//		Rdb: rdb,
//	}
//}

// RedisCache Redis 对 repo.Cache 的具体实现
type RedisCache struct {
	Rdb *redis.Client
}

func (rc *RedisCache) Put(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	err := rc.Rdb.Set(ctx, key, value, expire).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := rc.Rdb.Get(ctx, key).Result()
	if err != nil {
		return result, err
	}
	return result, nil
}

func (rc *RedisCache) HSet(ctx context.Context, key string, field string, value string) {
	rc.Rdb.HSet(ctx, key, field, value)
}

func (rc *RedisCache) HKeys(ctx context.Context, key string) ([]string, error) {
	result, err := rc.Rdb.HKeys(ctx, key).Result()
	return result, err
}

func (rc *RedisCache) Delete(ctx context.Context, keys []string) {
	rc.Rdb.Del(ctx, keys...)
}
