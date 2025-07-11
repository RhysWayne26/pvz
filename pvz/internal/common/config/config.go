package config

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"pvz-cli/internal/common/constants"
	"strconv"
	"strings"
)

const (
	dbMode   = "db"
	fileMode = "file"
)

// FileConfig holds the configuration for file-based storage, including the file path.
type FileConfig struct {
	Path string
}

// DBConfig holds the configuration for connecting to the database.
type DBConfig struct {
	WriteDSN string
	ReadDSN  string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

type OutboxConfig struct {
	BatchSize       int
	MaxAttempts     int
	RetryDelaySec   int
	PollIntervalSec int
}

// Config represents the application configuration, supporting both file-based and database-based configurations.
type Config struct {
	File   *FileConfig
	DB     *DBConfig
	Kafka  *KafkaConfig
	Outbox *OutboxConfig
}

// Load initializes and returns the application configuration based on environment variables and flags.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("failed to load .env, falling back to real environment", "error", err)
	}
	if os.Getenv("APP_ENV") == "test" {
		return loadTestConfig()
	}
	flag.Parse()
	mode := strings.TrimSpace(os.Getenv("STORAGE_MODE"))
	cfg := &Config{}
	switch mode {
	case dbMode:
		cfg.DB = loadDBConfig()
		cfg.Kafka = loadKafkaConfig()
		cfg.Outbox = loadOutboxConfig()
		validateKafkaOutbox(cfg)
	case fileMode:
		cfg.File = loadFileConfig()
		forbidKafkaOutboxInFileMode()
	default:
		slog.Error("invalid storage mode", "mode", mode)
		os.Exit(1)
	}
	return cfg
}

func loadDBConfig() *DBConfig {
	writeDSN := strings.TrimSpace(os.Getenv("DB_WRITE_DSN"))
	readDSN := strings.TrimSpace(os.Getenv("DB_READ_DSN"))
	if writeDSN == "" {
		user := strings.TrimSpace(os.Getenv("POSTGRES_USER"))
		pass := strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD"))
		host := strings.TrimSpace(os.Getenv("POSTGRES_HOST"))
		port := strings.TrimSpace(os.Getenv("POSTGRES_PORT"))
		db := strings.TrimSpace(os.Getenv("POSTGRES_DB"))
		if host == "" {
			host = constants.DefaultPGHost
		}
		if port == "" {
			port = constants.DefaultPGPort
		}
		if user == "" || pass == "" || db == "" {
			slog.Error("Missing required DB environment variables", "required", "POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB")
			os.Exit(1)
		}
		writeDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
	}
	if readDSN == "" {
		readDSN = writeDSN
	}
	return &DBConfig{
		WriteDSN: writeDSN,
		ReadDSN:  readDSN,
	}
}

func loadFileConfig() *FileConfig {
	path := strings.TrimSpace(os.Getenv("FILE_STORAGE_PATH"))
	if path == "" {
		path = constants.DefaultFileStoragePath
	}
	return &FileConfig{Path: path}
}

func loadTestConfig() *Config {
	testDSN := strings.TrimSpace(os.Getenv("TEST_DB_DSN"))
	if testDSN == "" {
		slog.Error("TEST_DB_DSN required when APP_ENV=test")
		os.Exit(1)
	}
	return &Config{
		DB: &DBConfig{
			WriteDSN: testDSN,
			ReadDSN:  testDSN,
		},
	}
}

func loadKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		Topic:   firstNonEmpty(os.Getenv("KAFKA_TOPIC"), "pvz.events-log"),
	}
}

func loadOutboxConfig() *OutboxConfig {
	return &OutboxConfig{
		BatchSize:       atoiDef(os.Getenv("OUTBOX_BATCH_SIZE"), 100),
		MaxAttempts:     atoiDef(os.Getenv("OUTBOX_MAX_ATTEMPTS"), 3),
		RetryDelaySec:   atoiDef(os.Getenv("OUTBOX_RETRY_DELAY_SEC"), 2),
		PollIntervalSec: atoiDef(os.Getenv("OUTBOX_POLL_INTERVAL_SEC"), 1),
	}
}
func validateKafkaOutbox(cfg *Config) {
	if len(cfg.Kafka.Brokers) == 0 || strings.TrimSpace(cfg.Kafka.Brokers[0]) == "" {
		slog.Error("KAFKA_BROKERS must be set when STORAGE_MODE=db")
		os.Exit(1)
	}
	if strings.TrimSpace(cfg.Kafka.Topic) == "" {
		slog.Error("KAFKA_TOPIC must be set when STORAGE_MODE=db")
		os.Exit(1)
	}
	if cfg.Outbox.BatchSize <= 0 {
		slog.Error("OUTBOX_BATCH_SIZE must be > 0 when STORAGE_MODE=db")
		os.Exit(1)
	}
	if cfg.Outbox.MaxAttempts <= 0 {
		slog.Error("OUTBOX_MAX_ATTEMPTS must be > 0 when STORAGE_MODE=db")
		os.Exit(1)
	}
	if cfg.Outbox.RetryDelaySec <= 0 {
		slog.Error("OUTBOX_RETRY_DELAY_SEC must be > 0 when STORAGE_MODE=db")
		os.Exit(1)
	}
	if cfg.Outbox.PollIntervalSec <= 0 {
		slog.Error("OUTBOX_POLL_INTERVAL_SEC must be > 0 when STORAGE_MODE=db")
		os.Exit(1)
	}
}

func forbidKafkaOutboxInFileMode() {
	if s := os.Getenv("KAFKA_BROKERS"); strings.TrimSpace(s) != "" {
		slog.Error("KAFKA_BROKERS must not be set when STORAGE_MODE=file", "value", s)
		os.Exit(1)
	}
	if s := os.Getenv("KAFKA_TOPIC"); strings.TrimSpace(s) != "" {
		slog.Error("KAFKA_TOPIC must not be set when STORAGE_MODE=file", "value", s)
		os.Exit(1)
	}
	if s := os.Getenv("OUTBOX_BATCH_SIZE"); strings.TrimSpace(s) != "" {
		slog.Error("OUTBOX_BATCH_SIZE must not be set when STORAGE_MODE=file", "value", s)
		os.Exit(1)
	}
	if s := os.Getenv("OUTBOX_MAX_ATTEMPTS"); strings.TrimSpace(s) != "" {
		slog.Error("OUTBOX_MAX_ATTEMPTS must not be set when STORAGE_MODE=file", "value", s)
		os.Exit(1)
	}
	if s := os.Getenv("OUTBOX_RETRY_DELAY_SEC"); strings.TrimSpace(s) != "" {
		slog.Error("OUTBOX_RETRY_DELAY_SEC must not be set when STORAGE_MODE=file", "value", s)
		os.Exit(1)
	}
	if s := os.Getenv("OUTBOX_POLL_INTERVAL_SEC"); strings.TrimSpace(s) != "" {
		slog.Error("OUTBOX_POLL_INTERVAL_SEC must not be set when STORAGE_MODE=file", "value", s)
		os.Exit(1)
	}
}

func firstNonEmpty(val, def string) string {
	if strings.TrimSpace(val) != "" {
		return val
	}
	return def
}

func atoiDef(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return def
}
