//go:generate minimock -g -i * -o mocks -s "_mock.go"
package repositories

import (
	"context"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// HistoryRepository handles persistence operations for order history entries
type HistoryRepository interface {
	Save(ctx context.Context, e models.HistoryEntry) error
	List(ctx context.Context, filter requests.OrderHistoryFilter) ([]models.HistoryEntry, int, error)
}
