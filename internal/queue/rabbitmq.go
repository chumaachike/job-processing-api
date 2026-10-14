package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chumaachike/job-processing-api/internal/job"
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

func (r *RabbitMQ) Enqueue(ctx context.Context, job job.JobMessage) error {
	body, err := json.Marshal(job)

	if err != nil {
		return err
	}

	return r.channel.PublishWithContext(ctx, "", r.queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	},
	)
}
