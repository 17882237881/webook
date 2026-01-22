package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/internal/domain"

	"github.com/redis/go-redis/v9"
)

// PostCache 帖子缓存接口（只缓存已发布帖子）
type PostCache interface {
	// Get 从缓存获取已发布帖子
	Get(ctx context.Context, id int64) (domain.Post, error)
	// Set 设置缓存
	Set(ctx context.Context, p domain.Post) error
	// Delete 删除缓存（帖子更新/删除时）
	Delete(ctx context.Context, id int64) error
}

type postCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

// NewPostCache 创建帖子缓存实例
func NewPostCache(client redis.Cmdable) PostCache {
	return &postCache{
		client:     client,
		expiration: time.Minute * 15, // 缓存15分钟
	}
}

func (c *postCache) key(id int64) string {
	return fmt.Sprintf("post:published:%d", id)
}

func (c *postCache) Get(ctx context.Context, id int64) (domain.Post, error) {
	data, err := c.client.Get(ctx, c.key(id)).Bytes()
	if err != nil {
		return domain.Post{}, err
	}
	var p domain.Post
	err = json.Unmarshal(data, &p)
	return p, err
}

func (c *postCache) Set(ctx context.Context, p domain.Post) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(p.Id), data, c.expiration).Err()
}

func (c *postCache) Delete(ctx context.Context, id int64) error {
	return c.client.Del(ctx, c.key(id)).Err()
}
