package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/internal/domain"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExist = redis.Nil

// UserCache 用户缓存接口
type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
	Delete(ctx context.Context, id int64) error
}

// RedisUserCache Redis 实现的用户缓存
type RedisUserCache struct {
	client     redis.Cmdable // Redis 客户端
	expiration time.Duration // 缓存过期时间
}

// NewUserCache 创建用户缓存实例
func NewUserCache(client redis.Cmdable, expiration time.Duration) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: expiration,
	}
}

// key 生成缓存 key
func (c *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

// Get 从缓存获取用户信息
func (c *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := c.key(id)
	val, err := c.client.Get(ctx, key).Bytes() // 获取缓存值
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u) // 反序列化
	return u, err
}

// Set 将用户信息写入缓存
func (c *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u) // 序列化
	if err != nil {
		return err
	}
	key := c.key(u.Id)
	return c.client.Set(ctx, key, val, c.expiration).Err() // 设置缓存
}

// Delete 删除用户缓存（用于密码修改等场景）
func (c *RedisUserCache) Delete(ctx context.Context, id int64) error {
	return c.client.Del(ctx, c.key(id)).Err()
}
