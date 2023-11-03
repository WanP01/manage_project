package repo

import (
	"context"
	"time"
)

// Cache 内存存储相关接口抽象，实现以下功能(以便后续实现不一样的缓存模块)
// 1. put 放入
// 2. get 取出
type Cache interface {
	Put(ctx context.Context, key string, value interface{}, expire time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	HSet(ctx context.Context, key string, field string, value string)
	HKeys(ctx context.Context, key string) ([]string, error)
	Delete(background context.Context, keys []string)
}
