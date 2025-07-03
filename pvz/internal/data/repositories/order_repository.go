//go:generate minimock -g -i * -o mocks -s "_mock.go"
package repositories

import (
	"context"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// OrderRepository handles persistence operations for orders
type OrderRepository interface {
	Save(ctx context.Context, order models.Order) error
	Load(ctx context.Context, id uint64) (models.Order, error)
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, int, error)
}
