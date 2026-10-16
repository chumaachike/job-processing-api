package job

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	request   CreateJobRequest
	filter    JobFilter
	status    JobStatus
	calls     int
	err       error
	lookupErr error
	message   JobMessage
}

func (f *fakeRepository) CreateJob(_ context.Context, r CreateJobRequest) (JobMessage, error) {
	f.calls++
	f.request = r
	return f.message, f.err
}
func (f *fakeRepository) ListJobs(_ context.Context, r JobFilter) ([]Job, error) {
	f.calls++
	f.filter = r
	return []Job{}, f.err
}
func (f *fakeRepository) GetJob(_ context.Context, id uuid.UUID) (Job, error) {
	f.calls++
	return Job{ID: id}, f.err
}
func (f *fakeRepository) UpdateJob(_ context.Context, id uuid.UUID, s JobStatus) (Job, error) {
	f.calls++
	f.status = s
	return Job{ID: id, Status: s}, f.err
}
func (f *fakeRepository) GetJobByIdemKey(context.Context, string) (JobMessage, error) {
	return f.message, f.lookupErr
}

type fakePublisher struct {
	calls   int
	err     error
	message JobMessage
}

func (f *fakePublisher) Publish(_ context.Context, m JobMessage) error {
	f.calls++
	f.message = m
	return f.err
}

func TestCreateJob(t *testing.T) {
	boom := errors.New("dependency failure")
	for _, tc := range []struct {
		name                                    string
		req                                     CreateJobRequest
		repoErr, lookupErr, publishErr, wantErr error
		creates, publishes                      int
	}{
		{name: "blank type", req: CreateJobRequest{Type: "  "}, wantErr: ErrInvalidJob},
		{name: "invalid payload", req: CreateJobRequest{Type: "email", Payload: json.RawMessage(`{`)}, wantErr: ErrInvalidJob},
		{name: "defaults", req: CreateJobRequest{Type: " email "}, creates: 1, publishes: 1},
		{name: "existing key", req: CreateJobRequest{Type: "email", IdempotencyKey: "key"}},
		{name: "new key", req: CreateJobRequest{Type: "email", IdempotencyKey: "key"}, lookupErr: ErrJobNotFound, creates: 1, publishes: 1},
		{name: "lookup error", req: CreateJobRequest{Type: "email", IdempotencyKey: "key"}, lookupErr: boom, wantErr: boom},
		{name: "insert error", req: CreateJobRequest{Type: "email"}, repoErr: boom, wantErr: boom, creates: 1},
		{name: "publish error", req: CreateJobRequest{Type: "email"}, publishErr: boom, wantErr: boom, creates: 1, publishes: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeRepository{err: tc.repoErr, lookupErr: tc.lookupErr, message: JobMessage{JobId: uuid.New(), Type: "email"}}
			p := &fakePublisher{err: tc.publishErr}
			got, err := NewService(r, p).CreateJob(context.Background(), tc.req)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if r.calls != tc.creates || p.calls != tc.publishes {
				t.Fatalf("create/publish calls = %d/%d", r.calls, p.calls)
			}
			if err == nil && got != r.message {
				t.Fatalf("message = %+v", got)
			}
			if tc.creates > 0 && (r.request.Type != "email" || string(r.request.Payload) != "{}") {
				t.Fatalf("normalized request = %+v", r.request)
			}
		})
	}
}

func TestValidationAndDelegation(t *testing.T) {
	ctx := context.Background()
	r := &fakeRepository{}
	s := NewService(r, &fakePublisher{})
	id := uuid.New().String()
	for _, status := range []JobStatus{"", JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed} {
		if _, err := s.ListJobs(ctx, JobFilter{Type: " email ", Status: JobStatus(" " + string(status) + " ")}); err != nil {
			t.Fatal(err)
		}
		if r.filter.Type != "email" || r.filter.Status != status {
			t.Fatalf("filter = %+v", r.filter)
		}
	}
	before := r.calls
	if _, err := s.ListJobs(ctx, JobFilter{Status: "unknown"}); !errors.Is(err, ErrInvalidJobFilter) {
		t.Fatal(err)
	}
	if _, err := s.GetJob(ctx, "bad"); !errors.Is(err, ErrInvalidJobID) {
		t.Fatal(err)
	}
	if _, err := s.UpdateJobStatus(ctx, "bad", JobStatusRunning); !errors.Is(err, ErrInvalidJobID) {
		t.Fatal(err)
	}
	if _, err := s.UpdateJobStatus(ctx, id, "unknown"); !errors.Is(err, ErrInvalidJobStatus) {
		t.Fatal(err)
	}
	if r.calls != before {
		t.Fatal("invalid requests reached repository")
	}
	if got, err := s.GetJob(ctx, " "+id+" "); err != nil || got.ID.String() != id {
		t.Fatalf("get = %+v, %v", got, err)
	}
	for _, status := range []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed} {
		if _, err := s.UpdateJobStatus(ctx, id, JobStatus(" "+string(status)+" ")); err != nil || r.status != status {
			t.Fatalf("update: %v", err)
		}
	}
	r.err = errors.New("database unavailable")
	if _, err := s.ListJobs(ctx, JobFilter{}); !errors.Is(err, r.err) {
		t.Fatal(err)
	}
	if _, err := s.GetJob(ctx, id); !errors.Is(err, r.err) {
		t.Fatal(err)
	}
	if _, err := s.UpdateJobStatus(ctx, id, JobStatusRunning); !errors.Is(err, r.err) {
		t.Fatal(err)
	}
}
