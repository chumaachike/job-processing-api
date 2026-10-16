package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInfrastructureRoutes(t *testing.T) {
	router := NewRouter(nil)
	for _, tc := range []struct {
		path     string
		status   int
		contains string
	}{
		{"/health/live", 200, `"status":"ok"`},
		{"/metrics", 200, "go_goroutines"},
		{"/missing", 404, "404"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
			if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.contains) {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if tc.path == "/health/live" && !json.Valid(w.Body.Bytes()) {
				t.Fatal("invalid health JSON")
			}
		})
	}
}
