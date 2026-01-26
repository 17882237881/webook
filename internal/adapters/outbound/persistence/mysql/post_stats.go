package mysql

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostStats stores counters for a post.
type PostStats struct {
	PostId     int64 `gorm:"primaryKey"`
	LikeCnt    int64
	CollectCnt int64
	ReadCnt    int64
	Ctime      int64
	Utime      int64
}

// PostLikeRelation stores like status per user and post.
type PostLikeRelation struct {
	Id     int64 `gorm:"primaryKey,autoIncrement"`
	PostId int64 `gorm:"uniqueIndex:idx_like_post_user"`
	UserId int64 `gorm:"uniqueIndex:idx_like_post_user"`
	Status uint8
	Ctime  int64
	Utime  int64
}

// PostCollectRelation stores collect status per user and post.
type PostCollectRelation struct {
	Id     int64 `gorm:"primaryKey,autoIncrement"`
	PostId int64 `gorm:"uniqueIndex:idx_collect_post_user"`
	UserId int64 `gorm:"uniqueIndex:idx_collect_post_user"`
	Status uint8
	Ctime  int64
	Utime  int64
}

type PostStatsDAO struct {
	db *gorm.DB
}

func NewPostStatsDAO(db *gorm.DB) *PostStatsDAO {
	return &PostStatsDAO{db: db}
}

func (dao *PostStatsDAO) FindByPostIds(ctx context.Context, postIds []int64) ([]PostStats, error) {
	var stats []PostStats
	err := dao.db.WithContext(ctx).Where("post_id IN ?", postIds).Find(&stats).Error
	return stats, err
}

func (dao *PostStatsDAO) Upsert(ctx context.Context, stats []PostStats) error {
	now := time.Now().UnixMilli()
	for i := range stats {
		if stats[i].Ctime == 0 {
			stats[i].Ctime = now
		}
		stats[i].Utime = now
	}
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "post_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"like_cnt", "collect_cnt", "read_cnt", "utime"}),
	}).Create(&stats).Error
}

type PostLikeDAO struct {
	db *gorm.DB
}

func NewPostLikeDAO(db *gorm.DB) *PostLikeDAO {
	return &PostLikeDAO{db: db}
}

func (dao *PostLikeDAO) FindByPostIdUserId(ctx context.Context, postId, userId int64) (PostLikeRelation, error) {
	var rel PostLikeRelation
	err := dao.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postId, userId).First(&rel).Error
	return rel, err
}

func (dao *PostLikeDAO) Insert(ctx context.Context, rel *PostLikeRelation) error {
	now := time.Now().UnixMilli()
	rel.Ctime = now
	rel.Utime = now
	return dao.db.WithContext(ctx).Create(rel).Error
}

func (dao *PostLikeDAO) UpdateStatus(ctx context.Context, id int64, status uint8) error {
	return dao.db.WithContext(ctx).Model(&PostLikeRelation{}).Where("id = ?", id).Updates(map[string]any{
		"status": status,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (dao *PostLikeDAO) FindByPostIds(ctx context.Context, postIds []int64, userId int64) ([]PostLikeRelation, error) {
	var rels []PostLikeRelation
	err := dao.db.WithContext(ctx).Where("user_id = ? AND post_id IN ?", userId, postIds).Find(&rels).Error
	return rels, err
}

type PostCollectDAO struct {
	db *gorm.DB
}

func NewPostCollectDAO(db *gorm.DB) *PostCollectDAO {
	return &PostCollectDAO{db: db}
}

func (dao *PostCollectDAO) FindByPostIdUserId(ctx context.Context, postId, userId int64) (PostCollectRelation, error) {
	var rel PostCollectRelation
	err := dao.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postId, userId).First(&rel).Error
	return rel, err
}

func (dao *PostCollectDAO) Insert(ctx context.Context, rel *PostCollectRelation) error {
	now := time.Now().UnixMilli()
	rel.Ctime = now
	rel.Utime = now
	return dao.db.WithContext(ctx).Create(rel).Error
}

func (dao *PostCollectDAO) UpdateStatus(ctx context.Context, id int64, status uint8) error {
	return dao.db.WithContext(ctx).Model(&PostCollectRelation{}).Where("id = ?", id).Updates(map[string]any{
		"status": status,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (dao *PostCollectDAO) FindByPostIds(ctx context.Context, postIds []int64, userId int64) ([]PostCollectRelation, error) {
	var rels []PostCollectRelation
	err := dao.db.WithContext(ctx).Where("user_id = ? AND post_id IN ?", userId, postIds).Find(&rels).Error
	return rels, err
}
