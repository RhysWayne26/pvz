package repositories

import (
	"github.com/google/uuid"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

type OrderRepository interface {
	Save(order models.Order) error
	Load(id uuid.UUID) (models.Order, error)
	Delete(id uuid.UUID) error
	List(filter requests.ListOrdersFilter) ([]models.Order, error)
}
