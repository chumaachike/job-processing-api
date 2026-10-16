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

func NewRabbitMQ(queueName, url string) (*RabbitMQ, error) {
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

func (r *RabbitMQ) Publish(ctx context.Context, job job.JobMessage) error {
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

func (r *RabbitMQ) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	if err := r.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		return nil, fmt.Errorf("set qos: %w", err)
	}

	deliveries, err := r.channel.ConsumeWithContext(
		ctx,
		r.queueName,
		"",    // consumer name
		false, // autoAck - IMPORTANT: false
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("consume jobs: %w", err)
	}

	return deliveries, nil
}
