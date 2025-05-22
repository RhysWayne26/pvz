package services

import (
	"github.com/google/uuid"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

type OrderService interface {
	AcceptOrder(req requests.AcceptOrderRequest) error
	IssueOrder(req requests.IssueOrderRequest) error
	ListOrders(filter requests.ListOrdersFilter) ([]models.Order, *uuid.UUID, int, error)
	ImportOrders(filePath string) (int, error)
}
