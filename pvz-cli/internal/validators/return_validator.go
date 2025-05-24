package validators

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

type ReturnValidator interface {
	ValidateClientReturn(orders []models.Order, req requests.ClientReturnRequest) error
	ValidateReturnToCourier(o models.Order) error
}
