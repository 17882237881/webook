package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"webook/internal/domain"
	ports "webook/internal/ports/output"

	"github.com/redis/go-redis/v9"
)

// RedisPostStatsCache caches post stats in Redis.
type RedisPostStatsCache struct {
	client redis.Cmdable
}

func NewPostStatsCache(client redis.Cmdable) ports.PostStatsCache {
	return &RedisPostStatsCache{client: client}
}

func (c *RedisPostStatsCache) key(postId int64) string {
	return fmt.Sprintf("post:stats:%d", postId)
}

func (c *RedisPostStatsCache) Get(ctx context.Context, postId int64) (domain.PostStats, error) {
	m, err := c.client.HGetAll(ctx, c.key(postId)).Result()
	if err != nil {
		return domain.PostStats{}, err
	}
	if len(m) == 0 {
		return domain.PostStats{}, ErrKeyNotExist
	}
	return parseStats(postId, m), nil
}

func (c *RedisPostStatsCache) BatchGet(ctx context.Context, postIds []int64) (map[int64]domain.PostStats, error) {
	if len(postIds) == 0 {
		return map[int64]domain.PostStats{}, nil
	}
	pipe := c.client.Pipeline()
	cmds := make(map[int64]*redis.MapStringStringCmd, len(postIds))
	for _, id := range postIds {
		cmds[id] = pipe.HGetAll(ctx, c.key(id))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[int64]domain.PostStats, len(postIds))
	for id, cmd := range cmds {
		m, err := cmd.Result()
		if err != nil {
			continue
		}
		if len(m) == 0 {
			continue
		}
		result[id] = parseStats(id, m)
	}
	return result, nil
}

func (c *RedisPostStatsCache) Set(ctx context.Context, stats domain.PostStats) error {
	return c.client.HSet(ctx, c.key(stats.PostId), map[string]any{
		"like_cnt":    stats.LikeCnt,
		"collect_cnt": stats.CollectCnt,
		"read_cnt":    stats.ReadCnt,
	}).Err()
}

func (c *RedisPostStatsCache) BatchSet(ctx context.Context, stats []domain.PostStats) error {
	if len(stats) == 0 {
		return nil
	}
	pipe := c.client.Pipeline()
	for _, st := range stats {
		pipe.HSet(ctx, c.key(st.PostId), map[string]any{
			"like_cnt":    st.LikeCnt,
			"collect_cnt": st.CollectCnt,
			"read_cnt":    st.ReadCnt,
		})
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisPostStatsCache) IncrLike(ctx context.Context, postId int64, delta int64) (int64, error) {
	return c.client.HIncrBy(ctx, c.key(postId), "like_cnt", delta).Result()
}

func (c *RedisPostStatsCache) IncrCollect(ctx context.Context, postId int64, delta int64) (int64, error) {
	return c.client.HIncrBy(ctx, c.key(postId), "collect_cnt", delta).Result()
}

func (c *RedisPostStatsCache) IncrRead(ctx context.Context, postId int64, delta int64) (int64, error) {
	return c.client.HIncrBy(ctx, c.key(postId), "read_cnt", delta).Result()
}

func (c *RedisPostStatsCache) MarkDirty(ctx context.Context, postId int64) error {
	return c.client.SAdd(ctx, "post:stats:dirty", postId).Err()
}

func (c *RedisPostStatsCache) PopDirty(ctx context.Context, count int64) ([]int64, error) {
	if count <= 0 {
		return nil, nil
	}
	vals, err := c.client.SPopN(ctx, "post:stats:dirty", count).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	result := make([]int64, 0, len(vals))
	for _, v := range vals {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		result = append(result, id)
	}
	return result, nil
}

func (c *RedisPostStatsCache) SetReadDedupe(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, 1, ttl).Result()
}

func (c *RedisPostStatsCache) SetEventProcessed(ctx context.Context, eventId string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("post:stats:event:%s", eventId)
	return c.client.SetNX(ctx, key, 1, ttl).Result()
}

func (c *RedisPostStatsCache) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, 1, ttl).Result()
}

func parseStats(postId int64, m map[string]string) domain.PostStats {
	return domain.PostStats{
		PostId:     postId,
		LikeCnt:    parseInt64(m["like_cnt"]),
		CollectCnt: parseInt64(m["collect_cnt"]),
		ReadCnt:    parseInt64(m["read_cnt"]),
	}
}

func parseInt64(val string) int64 {
	if val == "" {
		return 0
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
