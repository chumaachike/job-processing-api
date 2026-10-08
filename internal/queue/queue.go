package queue

import (
	"sync"

	"github.com/chumaachike/job-processing-api/internal/job"
)

type JobQueue struct {
	jobs chan job.JobMessage
	wg   sync.WaitGroup
}

func NewJobQUeue(capacity int) *JobQueue {
	return &JobQueue{
		jobs: make(chan job.JobMessage, capacity),
	}
}

func (q *JobQueue) TryEnqueue(job job.JobMessage) bool {
	select {
	case q.jobs <- job:
		return true
	default:
		return false
	}
}

func (q *JobQueue) Jobs() <-chan job.JobMessage {
	return q.jobs
}
