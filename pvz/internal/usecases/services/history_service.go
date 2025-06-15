package services

import (
	"context"
	"pvz-cli/internal/models"
)

// HistoryService handles order history operations and tracking
type HistoryService interface {
	Record(ctx context.Context, e models.HistoryEntry) error
	GetByOrder(ctx context.Context, orderID uint64) ([]models.HistoryEntry, error)
	ListAll(ctx context.Context, page, limit int) ([]models.HistoryEntry, error)
}
