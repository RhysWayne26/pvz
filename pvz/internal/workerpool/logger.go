package workerpool

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

func openLogFile(cfg *Config) (*slog.Logger, *os.File) {
	cfg.applyDefaults()
	if err := os.MkdirAll(filepath.Dir(cfg.LogPath), 0o755); err != nil {
		log.Fatalf("workerpool: cannot create log directory: %v", err)
	}
	f, err := os.OpenFile(cfg.LogPath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		log.Fatalf("workerpool: cannot open log file %q: %v", cfg.LogPath, err)
	}
	handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: cfg.LogLevel})
	logger := slog.New(handler)
	return logger, f
}
