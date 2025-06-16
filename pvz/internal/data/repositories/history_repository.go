package repositories

import (
	"context"
	"pvz-cli/internal/models"
)

// HistoryRepository handles persistence operations for order history entries
type HistoryRepository interface {
	Save(ctx context.Context, e models.HistoryEntry) error
	LoadByOrder(ctx context.Context, orderID uint64) ([]models.HistoryEntry, error)
	LoadAll(ctx context.Context, page, limit int) ([]models.HistoryEntry, error)
}
