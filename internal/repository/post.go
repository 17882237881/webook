package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

// PostRepository 帖子仓储接口（制作库 - 作者操作）
type PostRepository interface {
	// Create 创建帖子（草稿）
	Create(ctx context.Context, p domain.Post) (int64, error)

	// Update 更新帖子
	Update(ctx context.Context, p domain.Post) error

	// FindById 根据ID查找帖子（制作库）
	FindById(ctx context.Context, id int64) (domain.Post, error)

	// FindByAuthor 查找作者的帖子列表（分页）
	FindByAuthor(ctx context.Context, authorId int64, offset, limit int) ([]domain.Post, error)

	// Sync 发布：将制作库数据同步到线上库
	Sync(ctx context.Context, p domain.Post) (int64, error)

	// SyncStatus 同步状态（如设为仅自己可见，需同步删除线上库）
	SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error
}

// PublishedPostRepository 线上库仓储接口（读者阅读）
type PublishedPostRepository interface {
	// FindById 获取已发布帖子详情
	FindById(ctx context.Context, id int64) (domain.Post, error)

	// List 获取已发布帖子列表（分页）
	List(ctx context.Context, offset, limit int) ([]domain.Post, error)
}

// ============== PostRepository 实现 ==============

type postRepository struct {
	dao *dao.PostDAO
}

// NewPostRepository 创建帖子仓储实例
func NewPostRepository(dao *dao.PostDAO) PostRepository {
	return &postRepository{dao: dao}
}

func (r *postRepository) Create(ctx context.Context, p domain.Post) (int64, error) {
	return r.dao.Insert(ctx, r.toEntity(p))
}

func (r *postRepository) Update(ctx context.Context, p domain.Post) error {
	return r.dao.UpdateById(ctx, r.toEntity(p))
}

func (r *postRepository) FindById(ctx context.Context, id int64) (domain.Post, error) {
	p, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.Post{}, err
	}
	return r.toDomain(p), nil
}

func (r *postRepository) FindByAuthor(ctx context.Context, authorId int64, offset, limit int) ([]domain.Post, error) {
	posts, err := r.dao.FindByAuthor(ctx, authorId, offset, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Post, len(posts))
	for i, p := range posts {
		result[i] = r.toDomain(p)
	}
	return result, nil
}

func (r *postRepository) Sync(ctx context.Context, p domain.Post) (int64, error) {
	return r.dao.Sync(ctx, r.toEntity(p))
}

func (r *postRepository) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	return r.dao.SyncStatus(ctx, id, authorId, status)
}

// toEntity 将领域对象转换为 DAO 实体
func (r *postRepository) toEntity(p domain.Post) dao.Post {
	return dao.Post{
		Id:       p.Id,
		Title:    p.Title,
		Content:  p.Content,
		AuthorId: p.AuthorId,
		Status:   p.Status,
	}
}

// toDomain 将 DAO 实体转换为领域对象
func (r *postRepository) toDomain(p dao.Post) domain.Post {
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

// ============== PublishedPostRepository 实现 ==============

type publishedPostRepository struct {
	dao   *dao.PublishedPostDAO
	cache cache.PostCache
}

// NewPublishedPostRepository 创建线上库仓储实例
func NewPublishedPostRepository(dao *dao.PublishedPostDAO, c cache.PostCache) PublishedPostRepository {
	return &publishedPostRepository{dao: dao, cache: c}
}

func (r *publishedPostRepository) FindById(ctx context.Context, id int64) (domain.Post, error) {
	// 先查缓存
	p, err := r.cache.Get(ctx, id)
	if err == nil {
		return p, nil
	}
	// 缓存未命中，查数据库
	daoPost, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.Post{}, err
	}
	p = r.toDomain(daoPost)
	// 异步回写缓存（不阻塞主流程）
	go func() {
		_ = r.cache.Set(ctx, p)
	}()
	return p, nil
}

func (r *publishedPostRepository) List(ctx context.Context, offset, limit int) ([]domain.Post, error) {
	posts, err := r.dao.List(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Post, len(posts))
	for i, p := range posts {
		result[i] = r.toDomain(p)
	}
	return result, nil
}

// toDomain 将 DAO 实体转换为领域对象
func (r *publishedPostRepository) toDomain(p dao.PublishedPost) domain.Post {
	return domain.Post{
		Id:       p.Id,
		Title:    p.Title,
		Content:  p.Content,
		AuthorId: p.AuthorId,
		Status:   domain.PostStatusPublished, // 线上库都是已发布
		Ctime:    p.Ctime,
		Utime:    p.Utime,
	}
}
