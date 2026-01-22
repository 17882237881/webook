package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Post 制作库实体（作者编辑用）
type Post struct {
	Id       int64  `gorm:"primarykey,autoIncrement"`
	Title    string `gorm:"size:256"`
	Content  string `gorm:"type:text"`
	AuthorId int64  `gorm:"index"`
	Status   uint8  // 0-未发布，1-已发布，2-仅自己可见
	Ctime    int64
	Utime    int64
}

// PublishedPost 线上库实体（读者阅读用）
type PublishedPost struct {
	Id       int64  `gorm:"primarykey"` // 与 Post.Id 相同
	Title    string `gorm:"size:256"`
	Content  string `gorm:"type:text"`
	AuthorId int64  `gorm:"index"`
	Ctime    int64
	Utime    int64 // 发布时间
}

// PostDAO 帖子数据访问对象（制作库）
type PostDAO struct {
	db *gorm.DB
}

// NewPostDAO 创建 PostDAO 实例
func NewPostDAO(db *gorm.DB) *PostDAO {
	return &PostDAO{db: db}
}

// Insert 创建帖子
func (d *PostDAO) Insert(ctx context.Context, p Post) (int64, error) {
	now := time.Now().UnixMilli()
	p.Ctime = now
	p.Utime = now
	err := d.db.WithContext(ctx).Create(&p).Error
	return p.Id, err
}

// UpdateById 根据ID更新帖子（只能更新自己的帖子）
func (d *PostDAO) UpdateById(ctx context.Context, p Post) error {
	now := time.Now().UnixMilli()
	res := d.db.WithContext(ctx).Model(&Post{}).
		Where("id = ? AND author_id = ?", p.Id, p.AuthorId).
		Updates(map[string]any{
			"title":   p.Title,
			"content": p.Content,
			"status":  p.Status,
			"utime":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindById 根据ID查找帖子
func (d *PostDAO) FindById(ctx context.Context, id int64) (Post, error) {
	var p Post
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	return p, err
}

// FindByAuthor 查找作者的帖子列表
func (d *PostDAO) FindByAuthor(ctx context.Context, authorId int64, offset, limit int) ([]Post, error) {
	var posts []Post
	err := d.db.WithContext(ctx).
		Where("author_id = ?", authorId).
		Order("utime DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// Sync 发布帖子（同步到线上库，事务操作）
func (d *PostDAO) Sync(ctx context.Context, p Post) (int64, error) {
	var id int64
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		// 1. 先处理制作库
		if p.Id == 0 {
			// 新建帖子直接发布
			now := time.Now().UnixMilli()
			p.Ctime = now
			p.Utime = now
			p.Status = 1 // 已发布
			err = tx.Create(&p).Error
			if err != nil {
				return err
			}
			id = p.Id
		} else {
			// 更新现有帖子并设置为已发布
			err = tx.Model(&Post{}).
				Where("id = ? AND author_id = ?", p.Id, p.AuthorId).
				Updates(map[string]any{
					"title":   p.Title,
					"content": p.Content,
					"status":  1, // 已发布
					"utime":   time.Now().UnixMilli(),
				}).Error
			if err != nil {
				return err
			}
			id = p.Id
		}

		// 2. Upsert 到线上库
		now := time.Now().UnixMilli()
		pubPost := PublishedPost{
			Id:       id,
			Title:    p.Title,
			Content:  p.Content,
			AuthorId: p.AuthorId,
			Ctime:    now,
			Utime:    now,
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "content", "utime"}),
		}).Create(&pubPost).Error
	})
	return id, err
}

// SyncStatus 同步状态（设为仅自己可见时，需同时删除线上库）
func (d *PostDAO) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 更新制作库状态
		res := tx.Model(&Post{}).
			Where("id = ? AND author_id = ?", id, authorId).
			Updates(map[string]any{
				"status": status,
				"utime":  time.Now().UnixMilli(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// 2. 如果设为仅自己可见，删除线上库
		if status == 2 {
			return tx.Delete(&PublishedPost{}, "id = ?", id).Error
		}
		return nil
	})
}

// PublishedPostDAO 线上库数据访问对象（读者用）
type PublishedPostDAO struct {
	db *gorm.DB
}

// NewPublishedPostDAO 创建 PublishedPostDAO 实例
func NewPublishedPostDAO(db *gorm.DB) *PublishedPostDAO {
	return &PublishedPostDAO{db: db}
}

// FindById 获取已发布帖子
func (d *PublishedPostDAO) FindById(ctx context.Context, id int64) (PublishedPost, error) {
	var p PublishedPost
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	return p, err
}

// List 获取已发布帖子列表
func (d *PublishedPostDAO) List(ctx context.Context, offset, limit int) ([]PublishedPost, error) {
	var posts []PublishedPost
	err := d.db.WithContext(ctx).
		Order("utime DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	return posts, err
}
