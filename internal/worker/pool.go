package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/chumaachike/job-processing-api/internal/job"
	"github.com/chumaachike/job-processing-api/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Pool struct {
	workers int
	queue   *queue.RabbitMQ
}

func New(workers int, queue *queue.RabbitMQ) *Pool {
	return &Pool{
		workers: workers,
		queue:   queue,
	}
}

func (p *Pool) Start(ctx context.Context) error {
	deliveries, err := p.queue.Consume(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < p.workers; i++ {
		go p.worker(ctx, i, deliveries)
	}
	return nil
}

func (p *Pool) worker(ctx context.Context, id int, deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}

			var j job.JobMessage

			if err := json.Unmarshal(delivery.Body, &j); err != nil {
				slog.Error(
					"failed to decode job",
					"error", err,
				)

				_ = delivery.Nack(false, false)
				continue
			}

			if err := p.process(ctx, j); err != nil {
				slog.Error(
					"failed to process job",
					"job_id", j.JobId,
					"error", err,
				)

				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
		}
	}
}

func (p *Pool) process(ctx context.Context, j job.JobMessage) error {
	slog.Info(
		"processing job",
		"job_id", j.JobId,
		"type", j.Type,
	)

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-time.After(2 * time.Second):
	}

	slog.Info(
		"job completed",
		"job_id", j.JobId,
	)

	return nil
}
