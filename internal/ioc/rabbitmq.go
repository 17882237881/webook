package ioc

import (
	"webook/config"
	"webook/internal/adapters/outbound/mq"
	output "webook/internal/ports/output"
	"webook/pkg/logger"

	 amqp "github.com/rabbitmq/amqp091-go"
)

// NewRabbitMQConn creates a RabbitMQ connection.
func NewRabbitMQConn(cfg *config.Config) *amqp.Connection {
	conn, err := amqp.Dial(cfg.MQ.URL)
	if err != nil {
		panic(err)
	}
	return conn
}

func NewRabbitMQProducerChannel(conn *amqp.Connection) mq.ProducerChannel {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return mq.ProducerChannel{Channel: ch}
}

func NewRabbitMQConsumerChannel(conn *amqp.Connection) mq.ConsumerChannel {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return mq.ConsumerChannel{Channel: ch}
}

func NewPostStatsPublisher(ch mq.ProducerChannel, cfg *config.Config) output.PostStatsEventPublisher {
	pub, err := mq.NewRabbitMQStatsPublisher(ch, cfg.MQ.Exchange, cfg.MQ.Queue, cfg.MQ.RoutingKey)
	if err != nil {
		panic(err)
	}
	return pub
}

func NewPostStatsConsumer(ch mq.ConsumerChannel, cfg *config.Config, cache output.PostStatsCache, l logger.Logger) *mq.RabbitMQStatsConsumer {
	consumer, err := mq.NewRabbitMQStatsConsumer(ch, cfg.MQ.Exchange, cfg.MQ.Queue, cfg.MQ.RoutingKey, cfg.MQ.Prefetch, cache, l)
	if err != nil {
		panic(err)
	}
	return consumer
}
