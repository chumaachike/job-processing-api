package queue

import (
	"sync"

	"github.com/chumaachike/job-processing-api/internal/job"
)

type JobQueue struct {
	jobs chan job.Job
	wg   sync.WaitGroup
}

func NewJobQUeue(capacity int) *JobQueue {
	return &JobQueue{
		jobs: make(chan job.Job, capacity),
	}
}

func (q *JobQueue) TryEnqueue(job job.Job) bool {
	select {
	case q.jobs <- job:
		return true
	default:
		return false
	}
}

func (q *JobQueue) Jobs() <-chan job.Job {
	return q.jobs
}
