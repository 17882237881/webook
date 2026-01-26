package application

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"webook/internal/domain"
	input "webook/internal/ports/input"
	output "webook/internal/ports/output"

	"github.com/google/uuid"
)

type postInteractionService struct {
	likeRepo    output.PostLikeRepository
	collectRepo output.PostCollectRepository
	statsRepo   output.PostStatsRepository
	statsCache  output.PostStatsCache
	publisher   output.PostStatsEventPublisher
}

func NewPostInteractionService(
	likeRepo output.PostLikeRepository,
	collectRepo output.PostCollectRepository,
	statsRepo output.PostStatsRepository,
	statsCache output.PostStatsCache,
	publisher output.PostStatsEventPublisher,
) input.PostInteractionService {
	return &postInteractionService{
		likeRepo:    likeRepo,
		collectRepo: collectRepo,
		statsRepo:   statsRepo,
		statsCache:  statsCache,
		publisher:   publisher,
	}
}

func (s *postInteractionService) Like(ctx context.Context, postId, userId int64) error {
	changed, err := s.likeRepo.SetStatus(ctx, postId, userId, 1)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return s.publish(ctx, domain.PostStatsEventLike, postId, userId)
}

func (s *postInteractionService) Unlike(ctx context.Context, postId, userId int64) error {
	changed, err := s.likeRepo.SetStatus(ctx, postId, userId, 0)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return s.publish(ctx, domain.PostStatsEventUnlike, postId, userId)
}

func (s *postInteractionService) Collect(ctx context.Context, postId, userId int64) error {
	changed, err := s.collectRepo.SetStatus(ctx, postId, userId, 1)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return s.publish(ctx, domain.PostStatsEventCollect, postId, userId)
}

func (s *postInteractionService) Uncollect(ctx context.Context, postId, userId int64) error {
	changed, err := s.collectRepo.SetStatus(ctx, postId, userId, 0)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return s.publish(ctx, domain.PostStatsEventUncollect, postId, userId)
}

func (s *postInteractionService) Read(ctx context.Context, postId, userId int64, ip, userAgent string) error {
	key := s.readDedupeKey(postId, userId, ip, userAgent)
	ok, err := s.statsCache.SetReadDedupe(ctx, key, 30*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return s.publish(ctx, domain.PostStatsEventRead, postId, userId)
}

func (s *postInteractionService) GetStats(ctx context.Context, postId, userId int64) (domain.PostStats, domain.PostUserStats, error) {
	statsMap, userMap, err := s.GetStatsBatch(ctx, []int64{postId}, userId)
	if err != nil {
		return domain.PostStats{}, domain.PostUserStats{}, err
	}
	stats := statsMap[postId]
	userStats := userMap[postId]
	return stats, userStats, nil
}

func (s *postInteractionService) GetStatsBatch(ctx context.Context, postIds []int64, userId int64) (map[int64]domain.PostStats, map[int64]domain.PostUserStats, error) {
	if len(postIds) == 0 {
		return map[int64]domain.PostStats{}, map[int64]domain.PostUserStats{}, nil
	}
	stats := make(map[int64]domain.PostStats, len(postIds))
	userStats := make(map[int64]domain.PostUserStats, len(postIds))

	cacheStats, err := s.statsCache.BatchGet(ctx, postIds)
	if err != nil {
		cacheStats = map[int64]domain.PostStats{}
	}
	for id, st := range cacheStats {
		stats[id] = st
	}

	missing := make([]int64, 0, len(postIds))
	for _, id := range postIds {
		if _, ok := stats[id]; !ok {
			missing = append(missing, id)
		}
	}

	if len(missing) > 0 {
		dbStats, err := s.statsRepo.FindByPostIds(ctx, missing)
		if err != nil {
			return nil, nil, err
		}
		dbStatsMap := make(map[int64]domain.PostStats, len(dbStats))
		for _, st := range dbStats {
			stats[st.PostId] = st
			dbStatsMap[st.PostId] = st
		}
		for _, id := range missing {
			if _, ok := stats[id]; !ok {
				stats[id] = domain.PostStats{PostId: id}
			}
		}
		_ = s.statsCache.BatchSet(ctx, mapToSlice(stats, missing, dbStatsMap))
	}

	if userId > 0 {
		liked, err := s.likeRepo.FindLikedPostIds(ctx, postIds, userId)
		if err != nil && !errors.Is(err, context.Canceled) {
			return nil, nil, err
		}
		collected, err := s.collectRepo.FindCollectedPostIds(ctx, postIds, userId)
		if err != nil && !errors.Is(err, context.Canceled) {
			return nil, nil, err
		}
		for _, id := range postIds {
			userStats[id] = domain.PostUserStats{
				Liked:     liked[id],
				Collected: collected[id],
			}
		}
	} else {
		for _, id := range postIds {
			userStats[id] = domain.PostUserStats{}
		}
	}

	return stats, userStats, nil
}

func (s *postInteractionService) publish(ctx context.Context, eventType domain.PostStatsEventType, postId, userId int64) error {
	event := domain.NewPostStatsEvent(uuid.NewString(), eventType, postId, userId)
	return s.publisher.Publish(ctx, event)
}

func (s *postInteractionService) readDedupeKey(postId, userId int64, ip, userAgent string) string {
	if userId > 0 {
		return fmt.Sprintf("post:read:dedupe:uid:%d:pid:%d", userId, postId)
	}
	sum := sha1.Sum([]byte(ip + "|" + userAgent))
	return fmt.Sprintf("post:read:dedupe:anon:%s:pid:%d", hex.EncodeToString(sum[:]), postId)
}

func mapToSlice(all map[int64]domain.PostStats, missing []int64, db map[int64]domain.PostStats) []domain.PostStats {
	result := make([]domain.PostStats, 0, len(missing))
	for _, id := range missing {
		if st, ok := db[id]; ok {
			result = append(result, st)
		} else if st, ok := all[id]; ok {
			result = append(result, st)
		}
	}
	return result
}
