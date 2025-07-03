//go:generate minimock -g -i * -o mocks -s "_mock.go"
package services

import (
	"context"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// HistoryService handles order history operations and tracking
type HistoryService interface {
	Record(ctx context.Context, e models.HistoryEntry) error
	List(ctx context.Context, filter requests.OrderHistoryFilter) ([]models.HistoryEntry, error)
}
