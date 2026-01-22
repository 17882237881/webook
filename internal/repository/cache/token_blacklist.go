package cache

import (
	"context"
	"fmt"
	"time"
	"webook/internal/ports"

	"github.com/redis/go-redis/v9"
)

// RedisTokenBlacklist stores revoked tokens.
type RedisTokenBlacklist struct {
	client redis.Cmdable
}

func NewTokenBlacklist(client redis.Cmdable) ports.TokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

func (b *RedisTokenBlacklist) key(ssid string) string {
	return fmt.Sprintf("token:blacklist:%s", ssid)
}

func (b *RedisTokenBlacklist) Add(ctx context.Context, ssid string, expiration time.Duration) error {
	return b.client.Set(ctx, b.key(ssid), "1", expiration).Err()
}

func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, ssid string) (bool, error) {
	result, err := b.client.Exists(ctx, b.key(ssid)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
