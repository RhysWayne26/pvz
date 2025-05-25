package services

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/common"
	"pvz-cli/internal/usecases/requests"
)

type OrderService interface {
	AcceptOrder(req requests.AcceptOrderRequest) error
	IssueOrders(req requests.IssueOrdersRequest) []common.ProcessResult
	ListOrders(filter requests.ListOrdersFilter) ([]models.Order, string, int, error)
	ImportOrders(filePath string) (int, error)
}
