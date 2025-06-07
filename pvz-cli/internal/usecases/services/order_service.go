package services

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// OrderService handles certain order-related operation: acceptance, issuance, listing and returns
type OrderService interface {
	AcceptOrder(req requests.AcceptOrderRequest) (models.Order, error)
	IssueOrders(req requests.IssueOrdersRequest) []ProcessResult
	ListOrders(filter requests.ListOrdersRequest) ([]models.Order, uint64, int, error)
	CreateClientReturns(req requests.ClientReturnsRequest) []ProcessResult
	ReturnToCourier(req requests.ReturnOrderRequest) error
	ListReturns(page, limit int) ([]models.ReturnEntry, error)
}
