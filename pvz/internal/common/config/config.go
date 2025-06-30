package config

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"pvz-cli/internal/common/constants"
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

// Config represents the application configuration, supporting both file-based and database-based configurations.
type Config struct {
	File *FileConfig
	DB   *DBConfig
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
	switch mode {
	case dbMode:
		return &Config{DB: loadDBConfig()}
	case fileMode:
		return &Config{File: loadFileConfig()}
	default:
		slog.Error("invalid storage mode", "mode", mode)
		os.Exit(1)
		return nil
	}
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
