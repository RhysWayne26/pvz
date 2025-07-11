package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

// Config represents the configuration for connecting to a message broker.
type Config struct {
	Brokers []string
	GroupID string
	Topic   string
}

// Load reads configuration values from environment variables or a .env file and returns a Config struct or an error.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("failed to load .env, falling back to real environment", "error", err)
	}
	broker := os.Getenv("BROKER_ADDR")
	groupID := os.Getenv("GROUP_ID")
	topic := os.Getenv("TOPIC")

	if broker == "" {
		return nil, fmt.Errorf("BROKER_ADDR is required")
	}
	if groupID == "" {
		return nil, fmt.Errorf("GROUP_ID is required")
	}
	if topic == "" {
		return nil, fmt.Errorf("TOPIC is required")
	}

	return &Config{
		Brokers: []string{broker},
		GroupID: groupID,
		Topic:   topic,
	}, nil
}
