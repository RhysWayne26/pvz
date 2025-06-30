package services

import (
	"context"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services/shared"
)

// OrderService handles certain order-related operation: acceptance, issuance, listing and returns
type OrderService interface {
	AcceptOrder(ctx context.Context, req requests.AcceptOrderRequest) (models.Order, error)
	IssueOrders(ctx context.Context, req requests.IssueOrdersRequest) ([]shared.ProcessResult, error)
	ListOrders(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, uint64, int, error)
	CreateClientReturns(ctx context.Context, req requests.ClientReturnsRequest) ([]shared.ProcessResult, error)
	ReturnToCourier(ctx context.Context, req requests.ReturnOrderRequest) error
	ListReturns(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, error)
}
