package ioc

import (
	"webook/config"

	"github.com/redis/go-redis/v9"
)

// NewRedis 创建 Redis 客户端
func NewRedis(cfg *config.Config) redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})
}
