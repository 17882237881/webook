package application

import (
	"context"
	"webook/internal/domain"
	input "webook/internal/ports/input"
	output "webook/internal/ports/output"
)

type postService struct {
	repo    output.PostRepository
	pubRepo output.PublishedPostRepository
}

func NewPostService(repo output.PostRepository, pubRepo output.PublishedPostRepository) input.PostService {
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

func (s *postService) ListByAuthor(ctx context.Context, authorId int64, page, pageSize int) ([]domain.Post, int64, error) {
	offset := (page - 1) * pageSize
	posts, err := s.repo.FindByAuthor(ctx, authorId, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByAuthor(ctx, authorId)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (s *postService) GetPublishedById(ctx context.Context, id int64) (domain.Post, error) {
	return s.pubRepo.FindById(ctx, id)
}

func (s *postService) ListPublished(ctx context.Context, page, pageSize int) ([]domain.Post, int64, error) {
	offset := (page - 1) * pageSize
	posts, err := s.pubRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.pubRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}
