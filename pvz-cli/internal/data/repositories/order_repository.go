package repositories

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// OrderRepository handles persistence operations for orders
type OrderRepository interface {
	Save(order models.Order) error
	Load(id uint64) (models.Order, error)
	Delete(uint64) error
	List(filter requests.OrdersFilterRequest) ([]models.Order, int, error)
}
