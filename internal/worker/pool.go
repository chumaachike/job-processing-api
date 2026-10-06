package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/chumaachike/job-processing-api/internal/job"
	"github.com/chumaachike/job-processing-api/internal/queue"
)

type Pool struct {
	workers int
	queue   *queue.JobQueue
}

func New(workers int, queue *queue.JobQueue) *Pool {
	return &Pool{
		workers: workers,
		queue:   queue,
	}
}

func (p *Pool) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.queue.Jobs():
			if !ok {
				return
			}
			p.process(ctx, job)
		}
	}
}

func (p *Pool) process(ctx context.Context, j job.Job) {

	slog.Info(
		"processing job",
		"job_id", j.ID,
		"type", j.Type,
	)

	time.Sleep(2 * time.Second)

	slog.Info(
		"job completed",
		"job_id", j.ID,
	)
}
