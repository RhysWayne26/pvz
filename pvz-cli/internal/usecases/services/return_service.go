package services

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/common"
	"pvz-cli/internal/usecases/requests"
)

type ReturnService interface {
	CreateClientReturns(req requests.ClientReturnsRequest) []common.ProcessResult
	ReturnToCourier(req requests.ReturnOrderRequest) error
	ListReturns(page, limit int) ([]models.ReturnEntry, error)
}
