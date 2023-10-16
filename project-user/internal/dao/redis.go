package dao

import (
	"context"
	"project-user/config"
	"time"

	"github.com/go-redis/redis/v8"
)

var Rc *RedisCache

func init() {
	rdb := redis.NewClient(config.AppConf.InitRedisOptions())

	Rc = &RedisCache{
		rdb: rdb,
	}
}

// RedisCache Redis 对 repo.Cache 的具体实现
type RedisCache struct {
	rdb *redis.Client
}

func (rc *RedisCache) Put(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	err := rc.rdb.Set(ctx, key, value, expire).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := rc.rdb.Get(ctx, key).Result()
	if err != nil {
		return result, err
	}
	return result, nil
}
