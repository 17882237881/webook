package service

import (
	"context"
	"errors"
	"webook/internal/domain"
	"webook/internal/repository"
)

var (
	ErrPostNotFound         = errors.New("帖子不存在")
	ErrPostNotAuthor        = errors.New("无权操作该帖子")
	ErrPostAlreadyPublished = errors.New("帖子已发布")
)

// PostService 帖子服务接口
type PostService interface {
	// === 作者操作 ===

	// Save 保存帖子（创建或更新草稿）
	// 如果 p.Id == 0，则创建新帖子；否则更新现有帖子
	Save(ctx context.Context, p domain.Post) (int64, error)

	// Publish 发布帖子（同步到线上库）
	// 发布时会自动保存当前内容
	Publish(ctx context.Context, p domain.Post) (int64, error)

	// Delete 删除帖子（同时删除制作库和线上库）
	Delete(ctx context.Context, id int64, authorId int64) error

	// === 作者视角读取 ===

	// GetById 获取帖子详情（制作库，作者预览用）
	GetById(ctx context.Context, id int64) (domain.Post, error)

	// ListByAuthor 获取作者的帖子列表
	ListByAuthor(ctx context.Context, authorId int64, page, pageSize int) ([]domain.Post, error)

	// === 读者视角读取 ===

	// GetPublishedById 获取已发布帖子详情（线上库）
	GetPublishedById(ctx context.Context, id int64) (domain.Post, error)

	// ListPublished 获取已发布帖子列表（公开的）
	ListPublished(ctx context.Context, page, pageSize int) ([]domain.Post, error)
}

// postService 帖子服务实现
type postService struct {
	repo    repository.PostRepository
	pubRepo repository.PublishedPostRepository
}

// NewPostService 创建帖子服务实例
func NewPostService(repo repository.PostRepository, pubRepo repository.PublishedPostRepository) PostService {
	return &postService{
		repo:    repo,
		pubRepo: pubRepo,
	}
}

func (s *postService) Save(ctx context.Context, p domain.Post) (int64, error) {
	if p.Id == 0 {
		// 创建新帖子
		return s.repo.Create(ctx, p)
	}
	// 更新现有帖子
	return p.Id, s.repo.Update(ctx, p)
}

func (s *postService) Publish(ctx context.Context, p domain.Post) (int64, error) {
	// 发布时设置状态为已发布
	p.Status = domain.PostStatusPublished
	return s.repo.Sync(ctx, p)
}

func (s *postService) Delete(ctx context.Context, id int64, authorId int64) error {
	// 设置状态为仅自己可见（软删除），同时删除线上库
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
