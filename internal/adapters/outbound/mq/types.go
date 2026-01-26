package mq

import amqp"github.com/rabbitmq/amqp091-go"

type ProducerChannel struct {
	*amqp.Channel
}

type ConsumerChannel struct {
	*amqp.Channel
}
