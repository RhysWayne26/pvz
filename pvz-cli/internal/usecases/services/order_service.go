package services

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/common"
	"pvz-cli/internal/usecases/requests"
)

// OrderService handles certain order-related: acceptance, issuance and listing
type OrderService interface {
	AcceptOrder(req requests.AcceptOrderRequest) (models.Order, error)
	IssueOrders(req requests.IssueOrdersRequest) []common.ProcessResult
	ListOrders(filter requests.ListOrdersFilter) ([]models.Order, string, int, error)
}
