package repositories

import (
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/models"
)

type snapshotReturnRepository struct {
	storage storage.Storage
}

// NewSnapshotReturnRepository creates a new return repository using snapshot storage
func NewSnapshotReturnRepository(storage storage.Storage) ReturnRepository {
	return &snapshotReturnRepository{storage}
}

// Save stores a return entry in the repository
func (r *snapshotReturnRepository) Save(ret models.ReturnEntry) error {
	snap, err := r.storage.Load()
	if err != nil {
		return err
	}

	snap.Returns = append(snap.Returns, ret)
	return r.storage.Save(snap)
}

// List retrieves paginated list of return entries
func (r *snapshotReturnRepository) List(page, limit int) ([]models.ReturnEntry, error) {
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
