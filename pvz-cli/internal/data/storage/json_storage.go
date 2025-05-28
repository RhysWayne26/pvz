package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"pvz-cli/internal/data"
	"pvz-cli/internal/shutdown"
	"sync"
)

type JSONStorage struct {
	path  string
	mutex sync.Mutex
}

func NewJSONStorage(path string) *JSONStorage {
	return &JSONStorage{path: path}
}

func (s *JSONStorage) Save(snapshot *data.Snapshot) error {
	if shutdown.IsShuttingDown() {
		return fmt.Errorf("save aborted: shutting down")
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

func (s *JSONStorage) Load() (*data.Snapshot, error) {
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
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	var snapshot data.Snapshot
	if err := json.NewDecoder(file).Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	return &snapshot, nil
}
