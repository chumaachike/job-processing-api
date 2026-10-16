package database

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewPostgresIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	pool, err := NewPostgres(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if pool.Config().MaxConns != 10 || pool.Config().MinConns != 2 {
		t.Fatal("unexpected pool limits")
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestNewPostgresInvalidConfig(t *testing.T) {
	if pool, err := NewPostgres(context.Background(), "://invalid"); err == nil || pool != nil || !strings.Contains(err.Error(), "parse database config") {
		t.Fatalf("pool = %v, error = %v", pool, err)
	}
}
func TestNewPostgresCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if pool, err := NewPostgres(ctx, "postgres://user:pass@127.0.0.1:1/jobs?sslmode=disable"); err == nil || pool != nil {
		t.Fatalf("pool = %v, error = %v", pool, err)
	}
}
