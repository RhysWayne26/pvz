package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"pvz-cli/internal/data"
	"sync"
)

var _ Storage = (*JSONStorage)(nil)

// JSONStorage implements the storage interface using JSON files
type JSONStorage struct {
	path  string
	mutex sync.Mutex
}

// NewJSONStorage creates a new JSON-based storage with specified file path
func NewJSONStorage(path string) *JSONStorage {
	return &JSONStorage{path: path}
}

// Save persists snapshot to JSON file with atomic write operation
func (s *JSONStorage) Save(ctx context.Context, snapshot *data.Snapshot) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0750); err != nil {
		return fmt.Errorf("mkdir target dir: %w", err)
	}

	dir := filepath.Dir(s.path)
	pattern := filepath.Base(s.path) + ".tmp-*"
	tmpFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	cleanup := func() {
		err := tmpFile.Close()
		if err != nil {
			return
		}
		err = os.Remove(tmpPath)
		if err != nil {
			return
		}
	}

	defer func() {
		if _, err := os.Stat(tmpPath); err == nil {
			cleanup()
		}
	}()

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		cleanup()
		return fmt.Errorf("encode snapshot: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		cleanup()
		return fmt.Errorf("rename temp to target: %w", err)
	}

	return nil
}

// Load reads snapshot from JSON file or returns empty snapshot if file doesn't exist
func (s *JSONStorage) Load(ctx context.Context) (*data.Snapshot, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	file, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &data.Snapshot{}, nil
		}
		return nil, fmt.Errorf("open storage file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file: %v", err)
		}
	}()

	var snapshot data.Snapshot
	if err := json.NewDecoder(file).Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	return &snapshot, nil
}
