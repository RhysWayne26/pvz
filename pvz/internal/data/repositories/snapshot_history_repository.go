package repositories

import (
	"context"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"sort"
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

// List retrieves filtered and paginated history entries (by order ID or all)
func (r *SnapshotHistoryRepository) List(ctx context.Context, filter requests.OrderHistoryFilter) ([]models.HistoryEntry, int, error) {
	if ctx.Err() != nil {
		return nil, 0, ctx.Err()
	}
	snap, err := r.storage.Load(ctx)
	if err != nil {
		return nil, 0, err
	}
	var filtered []models.HistoryEntry
	if filter.OrderID != nil {
		for _, h := range snap.History {
			if h.OrderID == *filter.OrderID {
				filtered = append(filtered, h)
			}
		}
	} else {
		filtered = snap.History
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.Before(filtered[j].Timestamp)
	})
	start := (filter.Page - 1) * filter.Limit
	if start >= len(filtered) {
		return []models.HistoryEntry{}, 0, nil
	}
	end := start + filter.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], len(filtered), nil
}
