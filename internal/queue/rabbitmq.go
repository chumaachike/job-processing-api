package queue

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewRabbitMQ(url, queueName string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)

	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,  //durable
		false, // auto delete
		false, // exclusive
		false, // no wait
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}
	return &RabbitMQ{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}, nil

}
