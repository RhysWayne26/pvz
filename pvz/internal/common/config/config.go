package config

import (
	"flag"
	"fmt"
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
	DSN string
}

// Config represents the application configuration, supporting both file-based and database-based configurations.
type Config struct {
	File *FileConfig
	DB   *DBConfig
}

// Load initializes and returns the application configuration based on environment variables and flags.
func Load() *Config {
	flag.Parse()
	mode := strings.TrimSpace(os.Getenv("STORAGE_MODE"))
	switch mode {
	case dbMode:
		dsn := strings.TrimSpace(os.Getenv("DB_DSN"))
		if dsn == "" {
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
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
		}
		return &Config{DB: &DBConfig{DSN: dsn}}
	case fileMode:
		path := strings.TrimSpace(os.Getenv("FILE_STORAGE_PATH"))
		if path == "" {
			path = constants.DefaultFileStoragePath
		}
		return &Config{File: &FileConfig{Path: path}}
	default:
		fmt.Println("Invalid storage mode")
		os.Exit(1)
		return nil
	}
}
