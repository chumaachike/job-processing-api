package job

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Each test gets its own schema; never migrate or delete the public schema.
func TestPostgresRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := "test_" + uuid.New().String()
	schema = "\"" + schema + "\""
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.Exec(cleanupCtx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	}()
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	migrations, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	version := func(path string) int {
		n, err := strconv.Atoi(strings.SplitN(filepath.Base(path), "_", 2)[0])
		if err != nil {
			t.Fatalf("invalid migration version: %s", path)
		}
		return n
	}
	sort.Slice(migrations, func(i, j int) bool { return version(migrations[i]) < version(migrations[j]) })
	if len(migrations) != 3 {
		t.Fatalf("expected three migrations, got %d", len(migrations))
	}
	for _, path := range migrations {
		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("migration %s: %v", path, err)
		}
	}
	r := NewPostgresRepository(db, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := CreateJobRequest{Type: "email", Payload: json.RawMessage(`{"to":"test"}`)}
	first, err := r.CreateJob(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := r.CreateJob(ctx, request)
	if err != nil {
		t.Fatalf("multiple jobs without idempotency keys: %v", err)
	}
	if first.JobId == second.JobId || first.Type != "email" {
		t.Fatal("incorrect created messages")
	}
	got, err := r.GetJob(ctx, first.JobId)
	if err != nil || got.Status != JobStatusQueued || !json.Valid(got.Payload) || got.CreatedAt.IsZero() {
		t.Fatalf("get = %+v, %v", got, err)
	}
	updated, err := r.UpdateJob(ctx, first.JobId, JobStatusSucceeded)
	if err != nil || updated.Status != JobStatusSucceeded {
		t.Fatalf("update = %+v, %v", updated, err)
	}
	for _, tc := range []struct {
		filter JobFilter
		count  int
	}{{JobFilter{}, 2}, {JobFilter{Type: "email"}, 2}, {JobFilter{Status: JobStatusSucceeded}, 1}, {JobFilter{Type: "missing"}, 0}} {
		jobs, err := r.ListJobs(ctx, tc.filter)
		if err != nil || len(jobs) != tc.count {
			t.Fatalf("list %+v = %+v, %v", tc.filter, jobs, err)
		}
		for i := 1; i < len(jobs); i++ {
			if jobs[i].CreatedAt.After(jobs[i-1].CreatedAt) {
				t.Fatal("list order is not newest first")
			}
		}
	}
	request.IdempotencyKey = "unique-key"
	keyed, err := r.CreateJob(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	found, err := r.GetJobByIdemKey(ctx, request.IdempotencyKey)
	if err != nil || found != keyed {
		t.Fatalf("lookup = %+v, %v", found, err)
	}
	if _, err = r.CreateJob(ctx, request); err == nil {
		t.Fatal("duplicate key accepted")
	}
	if _, err = r.GetJob(ctx, uuid.New()); !errors.Is(err, ErrJobNotFound) {
		t.Fatal(err)
	}
	if _, err = r.UpdateJob(ctx, uuid.New(), JobStatusRunning); !errors.Is(err, ErrJobNotFound) {
		t.Fatal(err)
	}
	if _, err = r.GetJobByIdemKey(ctx, "missing"); !errors.Is(err, ErrJobNotFound) {
		t.Fatal(err)
	}
	if _, err = r.UpdateJob(ctx, first.JobId, "invalid"); err == nil {
		t.Fatal("database accepted invalid status")
	}
	canceled, stop := context.WithCancel(ctx)
	stop()
	if _, err = r.CreateJob(canceled, request); err == nil {
		t.Fatal("canceled insert succeeded")
	}
	if _, err = r.ListJobs(canceled, JobFilter{}); err == nil {
		t.Fatal("canceled list succeeded")
	}
	if _, err = r.GetJob(canceled, first.JobId); err == nil {
		t.Fatal("canceled get succeeded")
	}
	if _, err = r.UpdateJob(canceled, first.JobId, JobStatusRunning); err == nil {
		t.Fatal("canceled update succeeded")
	}
	if _, err = r.GetJobByIdemKey(canceled, "key"); err == nil {
		t.Fatal("canceled lookup succeeded")
	}
}
