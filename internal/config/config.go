package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	RabbitURL   string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	rabbitUser := os.Getenv("RABBIT_ADMIN")
	rabbitPassword := os.Getenv("RABBIT_PASSWORD")
	rabbitHost := os.Getenv("RABBIT_HOST")

	if rabbitHost == "" {
		rabbitHost = "localhost"
	}

	if rabbitUser == "" {
		return Config{}, fmt.Errorf("RABBIT_ADMIN is required")
	}

	if rabbitPassword == "" {
		return Config{}, fmt.Errorf("RABBIT_PASSWORD is required")
	}

	rabbitURI := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(rabbitUser, rabbitPassword),
		Host:   fmt.Sprintf("%s:5672", rabbitHost),
		Path:   "/",
	}

	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		RabbitURL:   rabbitURI.String(),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
