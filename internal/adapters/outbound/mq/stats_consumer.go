package mq

import (
	"context"
	"encoding/json"
	"time"
	"webook/internal/domain"
	output "webook/internal/ports/output"
	"webook/pkg/logger"

	 amqp"github.com/rabbitmq/amqp091-go"
)

type RabbitMQStatsConsumer struct {
	ch        *amqp.Channel
	queue     string
	prefetch  int
	cache     output.PostStatsCache
	logger    logger.Logger
	eventTTL  time.Duration
	closeChan chan struct{}
}

func NewRabbitMQStatsConsumer(
	ch ConsumerChannel,
	exchange, queue, routingKey string,
	prefetch int,
	cache output.PostStatsCache,
	l logger.Logger,
) (*RabbitMQStatsConsumer, error) {
	if err := ensureStatsTopology(ch.Channel, exchange, queue, routingKey); err != nil {
		return nil, err
	}
	return &RabbitMQStatsConsumer{
		ch:        ch.Channel,
		queue:     queue,
		prefetch:  prefetch,
		cache:     cache,
		logger:    l,
		eventTTL:  24 * time.Hour,
		closeChan: make(chan struct{}),
	}, nil
}

func (c *RabbitMQStatsConsumer) Start(ctx context.Context) {
	if c.prefetch > 0 {
		_ = c.ch.Qos(c.prefetch, 0, false)
	}
	msgs, err := c.ch.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		c.logger.Error("post stats consumer start failed", logger.Error(err))
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closeChan:
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			if err := c.handleMessage(ctx, msg); err != nil {
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

func (c *RabbitMQStatsConsumer) Stop() {
	close(c.closeChan)
}

func (c *RabbitMQStatsConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event domain.PostStatsEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.logger.Warn("post stats consumer invalid payload", logger.Error(err))
		return nil
	}

	ok, err := c.cache.SetEventProcessed(ctx, event.EventId, c.eventTTL)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	switch event.Type {
	case domain.PostStatsEventLike:
		_, err = c.cache.IncrLike(ctx, event.PostId, 1)
	case domain.PostStatsEventUnlike:
		_, err = c.cache.IncrLike(ctx, event.PostId, -1)
	case domain.PostStatsEventCollect:
		_, err = c.cache.IncrCollect(ctx, event.PostId, 1)
	case domain.PostStatsEventUncollect:
		_, err = c.cache.IncrCollect(ctx, event.PostId, -1)
	case domain.PostStatsEventRead:
		_, err = c.cache.IncrRead(ctx, event.PostId, 1)
	default:
		c.logger.Warn("post stats consumer unknown event type", logger.String("type", string(event.Type)))
		return nil
	}
	if err != nil {
		return err
	}
	return c.cache.MarkDirty(ctx, event.PostId)
}
