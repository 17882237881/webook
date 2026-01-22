package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/internal/domain"
	"webook/internal/ports"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExist = redis.Nil

// RedisUserCache is a Redis-backed cache.
type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

type UserCacheExpiration time.Duration

func NewUserCache(client redis.Cmdable, expiration UserCacheExpiration) ports.UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Duration(expiration),
	}
}

func (c *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func (c *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := c.key(id)
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (c *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := c.key(u.Id)
	return c.client.Set(ctx, key, val, c.expiration).Err()
}

func (c *RedisUserCache) Delete(ctx context.Context, id int64) error {
	return c.client.Del(ctx, c.key(id)).Err()
}
