package input

import (
	"context"
	"webook/internal/domain"
)

// PostService 帖子业务接口
type PostService interface {
	Save(ctx context.Context, post domain.Post) (int64, error)
	Publish(ctx context.Context, post domain.Post) (int64, error)
	GetById(ctx context.Context, id int64) (domain.Post, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Post, error)
	ListByAuthor(ctx context.Context, uid int64, page, pageSize int) ([]domain.Post, int64, error)
	ListPublished(ctx context.Context, page, pageSize int) ([]domain.Post, int64, error)
	Delete(ctx context.Context, id int64, uid int64) error
}
