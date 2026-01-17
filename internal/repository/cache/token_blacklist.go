package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklist Token 黑名单接口
// 用于退出登录时使 Refresh Token 失效
type TokenBlacklist interface {
	// Add 将 Token 加入黑名单
	// ssid: Token 的唯一标识（Session ID）
	// expiration: 黑名单记录的过期时间（应等于 Token 剩余有效期）
	Add(ctx context.Context, ssid string, expiration time.Duration) error

	// IsBlacklisted 检查 Token 是否在黑名单中
	IsBlacklisted(ctx context.Context, ssid string) (bool, error)
}

// RedisTokenBlacklist Redis 实现的 Token 黑名单
type RedisTokenBlacklist struct {
	client redis.Cmdable
}

// NewTokenBlacklist 创建 Token 黑名单实例
func NewTokenBlacklist(client redis.Cmdable) TokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

// key 生成黑名单 key
func (b *RedisTokenBlacklist) key(ssid string) string {
	return fmt.Sprintf("token:blacklist:%s", ssid)
}

// Add 将 Token 加入黑名单
func (b *RedisTokenBlacklist) Add(ctx context.Context, ssid string, expiration time.Duration) error {
	return b.client.Set(ctx, b.key(ssid), "1", expiration).Err()
}

// IsBlacklisted 检查 Token 是否在黑名单中
func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, ssid string) (bool, error) {
	result, err := b.client.Exists(ctx, b.key(ssid)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
