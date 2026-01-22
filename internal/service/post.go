package service

import (
	"context"
	"webook/internal/domain"
	"webook/internal/ports"
)

// PostService defines post use cases.
type PostService interface {
	Save(ctx context.Context, p domain.Post) (int64, error)
	Publish(ctx context.Context, p domain.Post) (int64, error)
	Delete(ctx context.Context, id int64, authorId int64) error
	GetById(ctx context.Context, id int64) (domain.Post, error)
	ListByAuthor(ctx context.Context, authorId int64, page, pageSize int) ([]domain.Post, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Post, error)
	ListPublished(ctx context.Context, page, pageSize int) ([]domain.Post, error)
}

type postService struct {
	repo    ports.PostRepository
	pubRepo ports.PublishedPostRepository
}

func NewPostService(repo ports.PostRepository, pubRepo ports.PublishedPostRepository) PostService {
	return &postService{
		repo:    repo,
		pubRepo: pubRepo,
	}
}

func (s *postService) Save(ctx context.Context, p domain.Post) (int64, error) {
	if p.Id == 0 {
		return s.repo.Create(ctx, p)
	}
	return p.Id, s.repo.Update(ctx, p)
}

func (s *postService) Publish(ctx context.Context, p domain.Post) (int64, error) {
	p.Status = domain.PostStatusPublished
	return s.repo.Sync(ctx, p)
}

func (s *postService) Delete(ctx context.Context, id int64, authorId int64) error {
	return s.repo.SyncStatus(ctx, id, authorId, domain.PostStatusPrivate)
}

func (s *postService) GetById(ctx context.Context, id int64) (domain.Post, error) {
	return s.repo.FindById(ctx, id)
}

func (s *postService) ListByAuthor(ctx context.Context, authorId int64, page, pageSize int) ([]domain.Post, error) {
	offset := (page - 1) * pageSize
	return s.repo.FindByAuthor(ctx, authorId, offset, pageSize)
}

func (s *postService) GetPublishedById(ctx context.Context, id int64) (domain.Post, error) {
	return s.pubRepo.FindById(ctx, id)
}

func (s *postService) ListPublished(ctx context.Context, page, pageSize int) ([]domain.Post, error) {
	offset := (page - 1) * pageSize
	return s.pubRepo.List(ctx, offset, pageSize)
}
