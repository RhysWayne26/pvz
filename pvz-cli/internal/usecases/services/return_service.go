package services

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

type ReturnService interface {
	CreateClientReturn(req requests.ClientReturnRequest) error
	ReturnToCourier(req requests.ReturnOrderRequest) error
	ListReturns(page, limit int) ([]models.ReturnEntry, error)
}
