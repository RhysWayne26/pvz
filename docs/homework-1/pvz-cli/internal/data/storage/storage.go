package storage

import (
	"pvz-cli/internal/data"
)

type Storage interface {
	Save(snapshot *data.Snapshot) error
	Load() (*data.Snapshot, error)
}
