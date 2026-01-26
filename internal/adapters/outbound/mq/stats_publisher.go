package mq

import (
	"context"
	"encoding/json"
	"webook/internal/domain"
	output "webook/internal/ports/output"

	amqp"github.com/rabbitmq/amqp091-go"
)

type RabbitMQStatsPublisher struct {
	ch         *amqp.Channel
	exchange   string
	routingKey string
}

func NewRabbitMQStatsPublisher(ch ProducerChannel, exchange, queue, routingKey string) (output.PostStatsEventPublisher, error) {
	if err := ensureStatsTopology(ch.Channel, exchange, queue, routingKey); err != nil {
		return nil, err
	}
	return &RabbitMQStatsPublisher{
		ch:         ch.Channel,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *RabbitMQStatsPublisher) Publish(ctx context.Context, event domain.PostStatsEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func ensureStatsTopology(ch *amqp.Channel, exchange, queue, routingKey string) error {
	if err := ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	return ch.QueueBind(q.Name, routingKey, exchange, false, nil)
}
