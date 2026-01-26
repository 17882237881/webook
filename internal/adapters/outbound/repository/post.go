package repository

import (
	"context"
	"errors"
	dao "webook/internal/adapters/outbound/persistence/mysql"
	"webook/internal/domain"
	ports "webook/internal/ports/output"

	"gorm.io/gorm"
)

// NewPostRepository builds a DAO-backed post repository.
func NewPostRepository(dao *dao.PostDAO) ports.PostRepository {
	return &postRepository{dao: dao}
}

// NewPublishedPostRepository builds a DAO-backed published repository.
func NewPublishedPostRepository(dao *dao.PublishedPostDAO) ports.PublishedPostRepository {
	return &publishedPostRepository{dao: dao}
}

// NewCachedPublishedPostRepository wraps a published repository with cache behavior.
func NewCachedPublishedPostRepository(repo ports.PublishedPostRepository, cache ports.PostCache) ports.PublishedPostRepository {
	return &cachedPublishedPostRepository{repo: repo, cache: cache}
}

type postRepository struct {
	dao *dao.PostDAO
}

type publishedPostRepository struct {
	dao *dao.PublishedPostDAO
}

type cachedPublishedPostRepository struct {
	repo  ports.PublishedPostRepository
	cache ports.PostCache
}

func (r *postRepository) Create(ctx context.Context, p domain.Post) (int64, error) {
	return r.dao.Insert(ctx, toPostEntity(p))
}

func (r *postRepository) Update(ctx context.Context, p domain.Post) error {
	err := r.dao.UpdateById(ctx, toPostEntity(p))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrPostNotFound
	}
	return err
}

func (r *postRepository) FindById(ctx context.Context, id int64) (domain.Post, error) {
	p, err := r.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Post{}, domain.ErrPostNotFound
		}
		return domain.Post{}, err
	}
	return toDomainPost(p), nil
}

func (r *postRepository) FindByAuthor(ctx context.Context, authorId int64, offset, limit int) ([]domain.Post, error) {
	posts, err := r.dao.FindByAuthor(ctx, authorId, offset, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Post, len(posts))
	for i, p := range posts {
		result[i] = toDomainPost(p)
	}
	return result, nil
}

func (r *postRepository) Sync(ctx context.Context, p domain.Post) (int64, error) {
	return r.dao.Sync(ctx, toPostEntity(p))
}

func (r *postRepository) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	err := r.dao.SyncStatus(ctx, id, authorId, status)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrPostNotFound
	}
	return err
}

func (r *postRepository) CountByAuthor(ctx context.Context, authorId int64) (int64, error) {
	return r.dao.CountByAuthor(ctx, authorId)
}

func (r *publishedPostRepository) FindById(ctx context.Context, id int64) (domain.Post, error) {
	p, err := r.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Post{}, domain.ErrPostNotFound
		}
		return domain.Post{}, err
	}
	return toDomainPublishedPost(p), nil
}

func (r *publishedPostRepository) List(ctx context.Context, offset, limit int) ([]domain.Post, error) {
	posts, err := r.dao.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Post, len(posts))
	for i, p := range posts {
		result[i] = toDomainPublishedPost(p)
	}
	return result, nil
}

func (r *publishedPostRepository) Count(ctx context.Context) (int64, error) {
	return r.dao.Count(ctx)
}

func (r *cachedPublishedPostRepository) FindById(ctx context.Context, id int64) (domain.Post, error) {
	p, err := r.cache.Get(ctx, id)
	if err == nil {
		return p, nil
	}

	p, err = r.repo.FindById(ctx, id)
	if err != nil {
		return domain.Post{}, err
	}
	go func() {
		_ = r.cache.Set(ctx, p)
	}()
	return p, nil
}

func (r *cachedPublishedPostRepository) List(ctx context.Context, offset, limit int) ([]domain.Post, error) {
	return r.repo.List(ctx, offset, limit)
}

func (r *cachedPublishedPostRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

func toPostEntity(p domain.Post) dao.Post {
	return dao.Post{
		Id:       p.Id,
		Title:    p.Title,
		Content:  p.Content,
		AuthorId: p.AuthorId,
		Status:   p.Status,
	}
}

func toDomainPost(p dao.Post) domain.Post {
	return domain.Post{
		Id:       p.Id,
		Title:    p.Title,
		Content:  p.Content,
		AuthorId: p.AuthorId,
		Status:   p.Status,
		Ctime:    p.Ctime,
		Utime:    p.Utime,
	}
}

func toDomainPublishedPost(p dao.PublishedPost) domain.Post {
	return domain.Post{
		Id:       p.Id,
		Title:    p.Title,
		Content:  p.Content,
		AuthorId: p.AuthorId,
		Status:   domain.PostStatusPublished,
		Ctime:    p.Ctime,
		Utime:    p.Utime,
	}
}
