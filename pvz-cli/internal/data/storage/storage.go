package storage

import (
	"pvz-cli/internal/data"
)

// Storage handles persistence operations for application snapshots
type Storage interface {
	Save(snapshot *data.Snapshot) error
	Load() (*data.Snapshot, error)
}
