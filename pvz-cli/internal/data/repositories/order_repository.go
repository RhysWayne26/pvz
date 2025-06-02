package repositories

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// OrderRepository handles persistence operations for orders
type OrderRepository interface {
	Save(order models.Order) error
	Load(id string) (models.Order, error)
	Delete(id string) error
	List(filter requests.ListOrdersFilter) ([]models.Order, int, error)
}
