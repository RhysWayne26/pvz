package repositories

import (
	"errors"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/models"
)

type snapshotHistoryRepository struct {
	storage storage.Storage
}

func NewSnapshotHistoryRepository(s storage.Storage) HistoryRepository {
	return &snapshotHistoryRepository{storage: s}
}

func (r *snapshotHistoryRepository) Save(e models.HistoryEntry) error {
	snap, err := r.storage.Load()
	if err != nil {
		return err
	}

	snap.History = append(snap.History, e)
	return r.storage.Save(snap)
}

func (r *snapshotHistoryRepository) LoadByOrder(orderID string) ([]models.HistoryEntry, error) {
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

func (r *snapshotHistoryRepository) LoadAll(page, limit int) ([]models.HistoryEntry, error) {
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
