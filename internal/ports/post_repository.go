package ports

import (
	"context"
	"webook/internal/domain"
)

type PostRepository interface {
	Create(ctx context.Context, p domain.Post) (int64, error)
	Update(ctx context.Context, p domain.Post) error
	FindById(ctx context.Context, id int64) (domain.Post, error)
	FindByAuthor(ctx context.Context, authorId int64, offset, limit int) ([]domain.Post, error)
	Sync(ctx context.Context, p domain.Post) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error
}

type PublishedPostRepository interface {
	FindById(ctx context.Context, id int64) (domain.Post, error)
	List(ctx context.Context, offset, limit int) ([]domain.Post, error)
}
