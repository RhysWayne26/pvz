package services

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

type OrderService interface {
	AcceptOrder(req requests.AcceptOrderRequest) error
	IssueOrder(req requests.IssueOrderRequest) error
	ListOrders(filter requests.ListOrdersFilter) ([]models.Order, string, int, error)
	ImportOrders(filePath string) (int, error)
}
