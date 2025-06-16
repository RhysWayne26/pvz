package storage

import (
	"context"
	"pvz-cli/internal/data"
)

// Storage handles persistence operations for application snapshots
type Storage interface {
	Save(ctx context.Context, snapshot *data.Snapshot) error
	Load(ctx context.Context) (*data.Snapshot, error)
}
