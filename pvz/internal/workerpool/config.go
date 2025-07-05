package workerpool

import (
	"log/slog"
	"time"
)

const (
	defaultWPLogPath     = "./logs/workerpool.log"
	defaultWorkerCount   = 4
	defaultStatsInterval = 30 * time.Second
	defaultQueueFactor   = 2
)

// Config defines the settings for configuring a worker pool, including logging, concurrency, and statistics options.
type Config struct {
	LogPath       string
	LogLevel      slog.Level
	WorkerCount   int
	StatsInterval time.Duration
	QueueFactor   int
}

func (c *Config) applyDefaults() {
	if c.LogPath == "" {
		c.LogPath = defaultWPLogPath
	}
	if c.LogLevel == 0 {
		c.LogLevel = slog.LevelInfo
	}
	if c.WorkerCount <= 0 {
		c.WorkerCount = defaultWorkerCount
	}
	if c.StatsInterval <= 0 {
		c.StatsInterval = defaultStatsInterval
	}
	if c.QueueFactor <= 0 {
		c.QueueFactor = defaultQueueFactor
	}
}

// Option defines a function that applies a configuration setting to a Config instance.
type Option func(*Config)

// WithLogPath sets the log file path in the Config structure. It applies this configuration using the provided Option.
func WithLogPath(path string) Option {
	return func(c *Config) {
		c.LogPath = path
	}
}

// WithLogLevel sets the logging level for the Config instance by applying the specified slog.Level.
func WithLogLevel(level slog.Level) Option {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// WithWorkerCount sets the WorkerCount in the configuration to the specified value.
func WithWorkerCount(n int) Option {
	return func(c *Config) {
		c.WorkerCount = n
	}
}

// WithStatsInterval sets the interval at which statistics are collected in the Config.
func WithStatsInterval(d time.Duration) Option {
	return func(c *Config) {
		c.StatsInterval = d
	}
}

// WithQueueFactor sets the QueueFactor configuration for a worker pool, determining the queue size multiplier for workers.
func WithQueueFactor(f int) Option {
	return func(c *Config) {
		c.QueueFactor = f
	}
}
