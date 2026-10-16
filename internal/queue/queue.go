package queue

import (
	"context"
	"errors"
	"sync"

	"github.com/chumaachike/job-processing-api/internal/job"
)

var ErrQueueFull = errors.New("job queue is full")

type JobQueue struct {
	jobs chan job.JobMessage
	wg   sync.WaitGroup
}

func NewJobQUeue(capacity int) *JobQueue {
	return &JobQueue{
		jobs: make(chan job.JobMessage, capacity),
	}
}

func (q *JobQueue) Publish(_ context.Context, job job.JobMessage) error {
	select {
	case q.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *JobQueue) Jobs() <-chan job.JobMessage {
	return q.jobs
}
