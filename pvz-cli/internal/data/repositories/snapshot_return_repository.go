package repositories

import (
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/models"
)

// SnapshotReturnRepository is an implementation of the ReturnRepository interface that uses snapshot storage.
type SnapshotReturnRepository struct {
	storage storage.Storage
}

// NewSnapshotReturnRepository creates a new instance of SnapshotReturnRepository
func NewSnapshotReturnRepository(storage storage.Storage) *SnapshotReturnRepository {
	return &SnapshotReturnRepository{storage}
}

// Save stores a return entry in the repository
func (r *SnapshotReturnRepository) Save(ret models.ReturnEntry) error {
	snap, err := r.storage.Load()
	if err != nil {
		return err
	}

	snap.Returns = append(snap.Returns, ret)
	return r.storage.Save(snap)
}

// List retrieves paginated list of return entries
func (r *SnapshotReturnRepository) List(page, limit int) ([]models.ReturnEntry, error) {
	snap, err := r.storage.Load()
	if err != nil {
		return nil, err
	}

	start := (page - 1) * limit
	if start >= len(snap.Returns) {
		return []models.ReturnEntry{}, nil
	}

	end := start + limit
	if end > len(snap.Returns) {
		end = len(snap.Returns)
	}

	return snap.Returns[start:end], nil
}
