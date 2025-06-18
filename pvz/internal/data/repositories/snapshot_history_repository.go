package repositories

import (
	"context"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/models"
)

var _ HistoryRepository = (*SnapshotHistoryRepository)(nil)

// SnapshotHistoryRepository is an implementation of the HistoryRepository interface that uses snapshot storage.
type SnapshotHistoryRepository struct {
	storage storage.Storage
}

// NewSnapshotHistoryRepository creates a new instance of SnapshotHistoryRepository
func NewSnapshotHistoryRepository(s storage.Storage) *SnapshotHistoryRepository {
	return &SnapshotHistoryRepository{storage: s}
}

// Save stores history entry in the repository
func (r *SnapshotHistoryRepository) Save(ctx context.Context, e models.HistoryEntry) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	snap, err := r.storage.Load(ctx)
	if err != nil {
		return err
	}

	snap.History = append(snap.History, e)
	return r.storage.Save(ctx, snap)
}

// LoadByOrder retrieves history entries for specific order
func (r *SnapshotHistoryRepository) LoadByOrder(ctx context.Context, orderID uint64) ([]models.HistoryEntry, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	snap, err := r.storage.Load(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []models.HistoryEntry
	for _, h := range snap.History {
		if h.OrderID == orderID {
			filtered = append(filtered, h)
		}
	}

	return filtered, nil
}

// LoadAll retrieves paginated list of all history entries
func (r *SnapshotHistoryRepository) LoadAll(ctx context.Context, page, limit int) ([]models.HistoryEntry, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	snap, err := r.storage.Load(ctx)
	if err != nil {
		return nil, err
	}

	start := (page - 1) * limit
	if start >= len(snap.History) {
		return nil, nil
	}

	end := start + limit
	if end > len(snap.History) {
		end = len(snap.History)
	}

	return snap.History[start:end], nil
}
