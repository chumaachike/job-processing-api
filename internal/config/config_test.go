package config

import (
	"net/url"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, tc := range []struct{ name, user, password, db, port, host, wantError string }{
		{name: "defaults", user: "guest", password: "guest", db: "postgres://localhost/jobs"},
		{name: "overrides", user: "u@ser", password: "p:/@ss", db: "postgres://localhost/jobs", port: "9000", host: "rabbit"},
		{name: "missing user", password: "pass", db: "db", wantError: "RABBIT_ADMIN"},
		{name: "missing password", user: "user", db: "db", wantError: "RABBIT_PASSWORD"},
		{name: "missing database", user: "user", password: "pass", wantError: "DATABASE_URL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range map[string]string{"RABBIT_ADMIN": tc.user, "RABBIT_PASSWORD": tc.password, "RABBIT_HOST": tc.host, "DATABASE_URL": tc.db, "PORT": tc.port} {
				t.Setenv(k, v)
			}
			cfg, err := Load()
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			port := tc.port
			if port == "" {
				port = "8080"
			}
			host := tc.host
			if host == "" {
				host = "localhost"
			}
			u, err := url.Parse(cfg.RabbitURL)
			if err != nil {
				t.Fatal(err)
			}
			password, _ := u.User.Password()
			if cfg.Port != port || cfg.DatabaseURL != tc.db || u.Host != host+":5672" || u.User.Username() != tc.user || password != tc.password {
				t.Fatalf("incorrect configuration: %+v", cfg)
			}
		})
	}
}
