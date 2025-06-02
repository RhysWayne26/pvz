package repositories

import (
	"errors"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/models"
)

// SnapshotHistoryRepository is an implementation of the HistoryRepository interface that uses snapshot storage.
type SnapshotHistoryRepository struct {
	storage storage.Storage
}

// NewSnapshotHistoryRepository creates a new instance of SnapshotHistoryRepository
func NewSnapshotHistoryRepository(s storage.Storage) *SnapshotHistoryRepository {
	return &SnapshotHistoryRepository{storage: s}
}

// Save stores history entry in the repository
func (r *SnapshotHistoryRepository) Save(e models.HistoryEntry) error {
	snap, err := r.storage.Load()
	if err != nil {
		return err
	}

	snap.History = append(snap.History, e)
	return r.storage.Save(snap)
}

// LoadByOrder retrieves history entries for specific order
func (r *SnapshotHistoryRepository) LoadByOrder(orderID string) ([]models.HistoryEntry, error) {
	snap, err := r.storage.Load()
	if err != nil {
		return nil, err
	}

	var filtered []models.HistoryEntry
	for _, h := range snap.History {
		if h.OrderID == orderID {
			filtered = append(filtered, h)
		}
	}

	if len(filtered) == 0 {
		return nil, errors.New("no history for this order")
	}

	return filtered, nil
}

// LoadAll retrieves paginated list of all history entries
func (r *SnapshotHistoryRepository) LoadAll(page, limit int) ([]models.HistoryEntry, error) {
	snap, err := r.storage.Load()
	if err != nil {
		return nil, err
	}

	start := (page - 1) * limit
	if start >= len(snap.History) {
		return []models.HistoryEntry{}, nil
	}

	end := start + limit
	if end > len(snap.History) {
		end = len(snap.History)
	}

	return snap.History[start:end], nil
}
