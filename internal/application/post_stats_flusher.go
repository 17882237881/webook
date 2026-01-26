package application

import (
	"context"
	"time"
	"webook/internal/domain"
	output "webook/internal/ports/output"
	"webook/pkg/logger"
)

type PostStatsFlusher struct {
	cache     output.PostStatsCache
	repo      output.PostStatsRepository
	logger    logger.Logger
	interval  time.Duration
	batchSize int64
	lockTTL   time.Duration
}

func NewPostStatsFlusher(cache output.PostStatsCache, repo output.PostStatsRepository, l logger.Logger) *PostStatsFlusher {
	return &PostStatsFlusher{
		cache:     cache,
		repo:      repo,
		logger:    l,
		interval:  5 * time.Second,
		batchSize: 100,
		lockTTL:   4 * time.Second,
	}
}

func (f *PostStatsFlusher) Start(ctx context.Context) {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f.FlushOnce(ctx)
		}
	}
}

func (f *PostStatsFlusher) FlushOnce(ctx context.Context) {
	locked, err := f.cache.TryLock(ctx, "post:stats:flush:lock", f.lockTTL)
	if err != nil || !locked {
		return
	}

	for {
		postIds, err := f.cache.PopDirty(ctx, f.batchSize)
		if err != nil {
			f.logger.Warn("post stats flush pop dirty failed", logger.Error(err))
			return
		}
		if len(postIds) == 0 {
			return
		}

		statsMap, err := f.cache.BatchGet(ctx, postIds)
		if err != nil {
			f.logger.Warn("post stats flush batch get failed", logger.Error(err))
			return
		}

		stats := make([]domain.PostStats, 0, len(statsMap))
		for _, st := range statsMap {
			stats = append(stats, st)
		}
		if len(stats) == 0 {
			return
		}
		if err := f.repo.Upsert(ctx, stats); err != nil {
			f.logger.Error("post stats flush upsert failed", logger.Error(err))
			return
		}
	}
}
