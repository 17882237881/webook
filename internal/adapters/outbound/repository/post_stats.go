package repository

import (
	"context"
	"errors"
	dao "webook/internal/adapters/outbound/persistence/mysql"
	"webook/internal/domain"
	output "webook/internal/ports/output"

	"gorm.io/gorm"
)

// NewPostStatsRepository builds a DAO-backed stats repository.
func NewPostStatsRepository(dao *dao.PostStatsDAO) output.PostStatsRepository {
	return &postStatsRepository{dao: dao}
}

// NewPostLikeRepository builds a DAO-backed like repository.
func NewPostLikeRepository(dao *dao.PostLikeDAO) output.PostLikeRepository {
	return &postLikeRepository{dao: dao}
}

// NewPostCollectRepository builds a DAO-backed collect repository.
func NewPostCollectRepository(dao *dao.PostCollectDAO) output.PostCollectRepository {
	return &postCollectRepository{dao: dao}
}

type postStatsRepository struct {
	dao *dao.PostStatsDAO
}

type postLikeRepository struct {
	dao *dao.PostLikeDAO
}

type postCollectRepository struct {
	dao *dao.PostCollectDAO
}

func (r *postStatsRepository) FindByPostIds(ctx context.Context, postIds []int64) ([]domain.PostStats, error) {
	stats, err := r.dao.FindByPostIds(ctx, postIds)
	if err != nil {
		return nil, err
	}
	result := make([]domain.PostStats, 0, len(stats))
	for _, st := range stats {
		result = append(result, domain.PostStats{
			PostId:     st.PostId,
			LikeCnt:    st.LikeCnt,
			CollectCnt: st.CollectCnt,
			ReadCnt:    st.ReadCnt,
		})
	}
	return result, nil
}

func (r *postStatsRepository) Upsert(ctx context.Context, stats []domain.PostStats) error {
	entities := make([]dao.PostStats, 0, len(stats))
	for _, st := range stats {
		entities = append(entities, dao.PostStats{
			PostId:     st.PostId,
			LikeCnt:    st.LikeCnt,
			CollectCnt: st.CollectCnt,
			ReadCnt:    st.ReadCnt,
		})
	}
	return r.dao.Upsert(ctx, entities)
}

func (r *postLikeRepository) SetStatus(ctx context.Context, postId, userId int64, status uint8) (bool, error) {
	rel, err := r.dao.FindByPostIdUserId(ctx, postId, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if status == 0 {
				return false, nil
			}
			newRel := &dao.PostLikeRelation{
				PostId: postId,
				UserId: userId,
				Status: status,
			}
			return true, r.dao.Insert(ctx, newRel)
		}
		return false, err
	}
	if rel.Status == status {
		return false, nil
	}
	return true, r.dao.UpdateStatus(ctx, rel.Id, status)
}

func (r *postLikeRepository) HasLiked(ctx context.Context, postId, userId int64) (bool, error) {
	rel, err := r.dao.FindByPostIdUserId(ctx, postId, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return rel.Status == 1, nil
}

func (r *postLikeRepository) FindLikedPostIds(ctx context.Context, postIds []int64, userId int64) (map[int64]bool, error) {
	rels, err := r.dao.FindByPostIds(ctx, postIds, userId)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]bool, len(rels))
	for _, rel := range rels {
		if rel.Status == 1 {
			result[rel.PostId] = true
		}
	}
	return result, nil
}

func (r *postCollectRepository) SetStatus(ctx context.Context, postId, userId int64, status uint8) (bool, error) {
	rel, err := r.dao.FindByPostIdUserId(ctx, postId, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if status == 0 {
				return false, nil
			}
			newRel := &dao.PostCollectRelation{
				PostId: postId,
				UserId: userId,
				Status: status,
			}
			return true, r.dao.Insert(ctx, newRel)
		}
		return false, err
	}
	if rel.Status == status {
		return false, nil
	}
	return true, r.dao.UpdateStatus(ctx, rel.Id, status)
}

func (r *postCollectRepository) HasCollected(ctx context.Context, postId, userId int64) (bool, error) {
	rel, err := r.dao.FindByPostIdUserId(ctx, postId, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return rel.Status == 1, nil
}

func (r *postCollectRepository) FindCollectedPostIds(ctx context.Context, postIds []int64, userId int64) (map[int64]bool, error) {
	rels, err := r.dao.FindByPostIds(ctx, postIds, userId)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]bool, len(rels))
	for _, rel := range rels {
		if rel.Status == 1 {
			result[rel.PostId] = true
		}
	}
	return result, nil
}
