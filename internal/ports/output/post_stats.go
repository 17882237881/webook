package output

import (
	"context"
	"time"
	"webook/internal/domain"
)

type PostStatsRepository interface {
	FindByPostIds(ctx context.Context, postIds []int64) ([]domain.PostStats, error)
	Upsert(ctx context.Context, stats []domain.PostStats) error
}

type PostStatsCache interface {
	Get(ctx context.Context, postId int64) (domain.PostStats, error)
	BatchGet(ctx context.Context, postIds []int64) (map[int64]domain.PostStats, error)
	Set(ctx context.Context, stats domain.PostStats) error
	BatchSet(ctx context.Context, stats []domain.PostStats) error

	IncrLike(ctx context.Context, postId int64, delta int64) (int64, error)
	IncrCollect(ctx context.Context, postId int64, delta int64) (int64, error)
	IncrRead(ctx context.Context, postId int64, delta int64) (int64, error)

	MarkDirty(ctx context.Context, postId int64) error
	PopDirty(ctx context.Context, count int64) ([]int64, error)

	SetReadDedupe(ctx context.Context, key string, ttl time.Duration) (bool, error)
	SetEventProcessed(ctx context.Context, eventId string, ttl time.Duration) (bool, error)
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

type PostLikeRepository interface {
	SetStatus(ctx context.Context, postId, userId int64, status uint8) (bool, error)
	HasLiked(ctx context.Context, postId, userId int64) (bool, error)
	FindLikedPostIds(ctx context.Context, postIds []int64, userId int64) (map[int64]bool, error)
}

type PostCollectRepository interface {
	SetStatus(ctx context.Context, postId, userId int64, status uint8) (bool, error)
	HasCollected(ctx context.Context, postId, userId int64) (bool, error)
	FindCollectedPostIds(ctx context.Context, postIds []int64, userId int64) (map[int64]bool, error)
}

type PostStatsEventPublisher interface {
	Publish(ctx context.Context, event domain.PostStatsEvent) error
}
