package job

import (
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	appmetrics "github.com/chumaachike/job-processing-api/metrics"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestRoutes(t *testing.T) {
	id := uuid.New().String()
	boom := errors.New("database offline")
	for _, tc := range []struct {
		name, method, path, body string
		repoErr, publishErr      error
		want                     int
	}{
		{name: "create", method: "POST", path: "/", body: `{"type":"email"}`, want: 201},
		{name: "unknown field", method: "POST", path: "/", body: `{"type":"email","extra":1}`, want: 400},
		{name: "invalid JSON", method: "POST", path: "/", body: `{`, want: 400},
		{name: "multiple objects", method: "POST", path: "/", body: `{"type":"email"} {}`, want: 400},
		{name: "missing type", method: "POST", path: "/", body: `{}`, want: 400},
		{name: "queue full", method: "POST", path: "/", body: `{"type":"email"}`, publishErr: ErrQueueFull, want: 503},
		{name: "create failure", method: "POST", path: "/", body: `{"type":"email"}`, repoErr: boom, want: 500},
		{name: "list", method: "GET", path: "/?type=email&status=queued", want: 200},
		{name: "invalid filter", method: "GET", path: "/?status=unknown", want: 400},
		{name: "list failure", method: "GET", path: "/", repoErr: boom, want: 500},
		{name: "get", method: "GET", path: "/" + id, want: 200},
		{name: "invalid id", method: "GET", path: "/bad", want: 400},
		{name: "missing job", method: "GET", path: "/" + id, repoErr: ErrJobNotFound, want: 404},
		{name: "get failure", method: "GET", path: "/" + id, repoErr: boom, want: 500},
		{name: "update", method: "PATCH", path: "/" + id, body: `{"status":"running"}`, want: 200},
		{name: "bad update id", method: "PATCH", path: "/bad", body: `{"status":"running"}`, want: 400},
		{name: "bad status", method: "PATCH", path: "/" + id, body: `{"status":"unknown"}`, want: 400},
		{name: "bad update JSON", method: "PATCH", path: "/" + id, body: `{`, want: 400},
		{name: "unknown update field", method: "PATCH", path: "/" + id, body: `{"extra":1}`, want: 400},
		{name: "trailing update JSON", method: "PATCH", path: "/" + id, body: `{"status":"running"} {}`, want: 400},
		{name: "missing update", method: "PATCH", path: "/" + id, body: `{"status":"running"}`, repoErr: ErrJobNotFound, want: 404},
		{name: "update failure", method: "PATCH", path: "/" + id, body: `{"status":"running"}`, repoErr: boom, want: 500},
		{name: "oversize create", method: "POST", path: "/", body: `{"type":"` + strings.Repeat("x", 1<<20) + `"}`, want: 400},
		{name: "oversize update", method: "PATCH", path: "/" + id, body: `{"status":"` + strings.Repeat("x", 1<<20) + `"}`, want: 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			counter := prometheus.NewCounter(prometheus.CounterOpts{Name: "test_created_total"})
			r := &fakeRepository{err: tc.repoErr, message: JobMessage{JobId: uuid.MustParse(id), Type: "email"}}
			h := NewHandler(NewService(r, &fakePublisher{err: tc.publishErr}), &appmetrics.Metrics{JobCreated: counter}, slog.New(slog.NewTextHandler(io.Discard, nil)))
			w := httptest.NewRecorder()
			Routes(h).ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.want, w.Body.String())
			}
			if tc.want >= 400 && strings.Contains(w.Body.String(), `"id"`) {
				t.Fatal("error response contains success body")
			}
			if tc.want < 300 && w.Header().Get("Content-Type") != "application/json" {
				t.Fatal("missing JSON content type")
			}
			metric := &dto.Metric{}
			if err := counter.Write(metric); err != nil {
				t.Fatal(err)
			}
			wantCount := float64(0)
			if tc.method == "POST" && tc.want == 201 {
				wantCount = 1
			}
			if metric.GetCounter().GetValue() != wantCount {
				t.Fatal("incorrect create metric")
			}
		})
	}
}
