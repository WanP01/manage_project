package config

import (
	"github.com/go-redis/redis/v8"
	"project-project/internal/dao"
)

func (c *Config) ReConnRedis() {
	rdb := redis.NewClient(c.InitRedisOptions())
	rc := &dao.RedisCache{
		Rdb: rdb,
	}
	dao.Rc = rc
}
