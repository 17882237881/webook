package application

import (
	"context"
)

type PostStatsWorker struct {
	consumer RabbitMQStatsConsumerWrapper
	flusher  *PostStatsFlusher
}

// RabbitMQStatsConsumerWrapper wraps a consumer without exposing MQ package to main.
type RabbitMQStatsConsumerWrapper interface {
	Start(ctx context.Context)
}

func NewPostStatsWorker(consumer RabbitMQStatsConsumerWrapper, flusher *PostStatsFlusher) *PostStatsWorker {
	return &PostStatsWorker{
		consumer: consumer,
		flusher:  flusher,
	}
}

func (w *PostStatsWorker) Start(ctx context.Context) {
	go w.consumer.Start(ctx)
	go w.flusher.Start(ctx)
}
