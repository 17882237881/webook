package input

import (
	"context"
	"webook/internal/domain"
)

// PostInteractionService handles likes/collects/reads and stats queries.
type PostInteractionService interface {
	Like(ctx context.Context, postId, userId int64) error
	Unlike(ctx context.Context, postId, userId int64) error
	Collect(ctx context.Context, postId, userId int64) error
	Uncollect(ctx context.Context, postId, userId int64) error
	Read(ctx context.Context, postId, userId int64, ip, userAgent string) error

	GetStats(ctx context.Context, postId, userId int64) (domain.PostStats, domain.PostUserStats, error)
	GetStatsBatch(ctx context.Context, postIds []int64, userId int64) (map[int64]domain.PostStats, map[int64]domain.PostUserStats, error)
}
