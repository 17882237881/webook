package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/internal/domain"
	ports "webook/internal/ports/output"

	"github.com/redis/go-redis/v9"
)

// RedisPostCache caches published posts.
type RedisPostCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewPostCache(client redis.Cmdable) ports.PostCache {
	return &RedisPostCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (c *RedisPostCache) key(id int64) string {
	return fmt.Sprintf("post:published:%d", id)
}

func (c *RedisPostCache) Get(ctx context.Context, id int64) (domain.Post, error) {
	data, err := c.client.Get(ctx, c.key(id)).Bytes()
	if err != nil {
		return domain.Post{}, err
	}
	var p domain.Post
	err = json.Unmarshal(data, &p)
	return p, err
}

func (c *RedisPostCache) Set(ctx context.Context, p domain.Post) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(p.Id), data, c.expiration).Err()
}

func (c *RedisPostCache) Delete(ctx context.Context, id int64) error {
	return c.client.Del(ctx, c.key(id)).Err()
}
