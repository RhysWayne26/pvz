package services

import (
	"context"
	"fmt"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services/validators"
	"time"
)

var _ OrderService = (*DefaultOrderService)(nil)

// DefaultOrderService is a default implementation of the OrderService interface
type DefaultOrderService struct {
	orderRepo         repositories.OrderRepository
	packagePricingSvc PackagePricingService
	historySvc        HistoryService
	validator         validators.OrderValidator
}

// NewDefaultOrderService creates a new instance of DefaultOrderService
func NewDefaultOrderService(
	orderRepo repositories.OrderRepository,
	packagePricingService PackagePricingService,
	historyService HistoryService,
	validator validators.OrderValidator) *DefaultOrderService {
	return &DefaultOrderService{
		orderRepo:         orderRepo,
		packagePricingSvc: packagePricingService,
		historySvc:        historyService,
		validator:         validator,
	}
}

// AcceptOrder accepts an order with package pricing calculation and validation
func (s *DefaultOrderService) AcceptOrder(ctx context.Context, req requests.AcceptOrderRequest) (models.Order, error) {
	if ctx.Err() != nil {
		return models.Order{}, ctx.Err()
	}
	existing, err := s.orderRepo.Load(ctx, req.OrderID)
	if err != nil {
		existing = models.Order{}
	}

	if err := s.validator.ValidateAccept(existing, req); err != nil {
		return models.Order{}, err
	}

	totalPrice, err := s.packagePricingSvc.Evaluate(req.Package, req.Weight, req.Price)
	if err != nil {
		return models.Order{}, err
	}

	now := time.Now()

	order := models.Order{
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		CreatedAt:       now,
		UpdatedStatusAt: now,
		Status:          models.Accepted,
		ExpiresAt:       req.ExpiresAt,
		Weight:          req.Weight,
		Price:           totalPrice,
		Package:         req.Package,
	}
	if err := s.orderRepo.Save(ctx, order); err != nil {
		return models.Order{}, apperrors.Newf(apperrors.InternalError, "failed to save order: %v", err)
	}

	entry := models.HistoryEntry{
		OrderID:   order.OrderID,
		Event:     models.EventAccepted,
		Timestamp: now,
	}
	if err := s.historySvc.Record(ctx, entry); err != nil {
		return models.Order{}, apperrors.Newf(apperrors.InternalError, "failed to record history: %v", err)
	}

	return order, nil
}

// IssueOrders processes multiple orders for issuance to clients
func (s *DefaultOrderService) IssueOrders(ctx context.Context, req requests.IssueOrdersRequest) ([]ProcessResult, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	results := make([]ProcessResult, 0, len(req.OrderIDs))
	now := time.Now()

	for _, id := range req.OrderIDs {
		res := ProcessResult{OrderID: id}

		order, err := s.orderRepo.Load(ctx, id)
		if err != nil {
			res.Error = apperrors.Newf(apperrors.OrderNotFound, "order %d not found", id)
			results = append(results, res)
			continue
		}

		if err := s.validator.ValidateIssue([]models.Order{order}, req); err != nil {
			res.Error = err
			results = append(results, res)
			continue
		}

		order.Status = models.Issued
		order.UpdatedStatusAt = now

		if err := s.orderRepo.Save(ctx, order); err != nil {
			res.Error = apperrors.Newf(apperrors.InternalError, "failed to save order %d: %v", id, err)
			results = append(results, res)
			continue
		}

		entry := models.HistoryEntry{
			OrderID:   order.OrderID,
			Event:     models.EventIssued,
			Timestamp: now,
		}

		if err := s.historySvc.Record(ctx, entry); err != nil {
			fmt.Printf("WARNING: failed to record history for order %d: %v\n", id, err)
		}

		res.Error = nil
		results = append(results, res)
	}

	return results, nil
}

// ListOrders retrieves filtered and paginated list of orders
func (s *DefaultOrderService) ListOrders(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, uint64, int, error) {
	result, total, err := s.orderRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, 0, apperrors.Newf(apperrors.InternalError, "failed to list orders: %v", err)
	}

	if filter.Last != nil {
		n := *filter.Last
		if len(result) > n {
			result = result[len(result)-n:]
		}
		total = len(result)
	}

	var nextLastID uint64
	if len(result) > 0 {
		nextLastID = result[len(result)-1].OrderID
	}

	return result, nextLastID, total, nil
}

// CreateClientReturns processes multiple client return requests
func (s *DefaultOrderService) CreateClientReturns(ctx context.Context, req requests.ClientReturnsRequest) ([]ProcessResult, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	results := make([]ProcessResult, 0, len(req.OrderIDs))
	now := time.Now()

	for _, id := range req.OrderIDs {
		res := ProcessResult{OrderID: id}

		order, err := s.orderRepo.Load(ctx, id)
		if err != nil {
			res.Error = apperrors.Newf(apperrors.OrderNotFound, "order %d not found", id)
			results = append(results, res)
			continue
		}

		if err := s.validator.ValidateClientReturn([]models.Order{order}, req); err != nil {
			res.Error = err
			results = append(results, res)
			continue
		}

		order.Status = models.Returned
		order.UpdatedStatusAt = now

		if err := s.orderRepo.Save(ctx, order); err != nil {
			res.Error = apperrors.Newf(apperrors.InternalError, "failed to save order %d: %v", order.OrderID, err)
			results = append(results, res)
			continue
		}

		entry := models.HistoryEntry{
			OrderID:   order.OrderID,
			Event:     models.EventReturnedFromClient,
			Timestamp: now,
		}
		if err := s.historySvc.Record(ctx, entry); err != nil {
			fmt.Printf("WARNING: failed to record history for order %d: %v\n", order.OrderID, err)
		}

		res.Error = nil
		results = append(results, res)
	}

	return results, nil
}

// ReturnToCourier processes return of order back to courier/warehouse
func (s *DefaultOrderService) ReturnToCourier(ctx context.Context, req requests.ReturnOrderRequest) error {
	orderID := req.OrderID
	o, err := s.orderRepo.Load(ctx, orderID)
	if err != nil {
		return apperrors.Newf(apperrors.OrderNotFound, "order %d not found", orderID)
	}

	if err := s.validator.ValidateReturnToCourier(o); err != nil {
		return err
	}

	if err := s.orderRepo.Delete(ctx, orderID); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to delete order %d: %v", orderID, err)
	}

	entry := models.HistoryEntry{
		OrderID:   orderID,
		Event:     models.EventReturnedToWarehouse,
		Timestamp: time.Now(),
	}
	if err := s.historySvc.Record(ctx, entry); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to record history for order %d: %v", orderID, err)
	}

	return nil
}

// ListReturns retrieves paginated list of return entries sorted by return date
func (s *DefaultOrderService) ListReturns(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, error) {
	orders, _, err := s.orderRepo.List(ctx, filter)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to list returns: %v", err)
	}

	return orders, nil
}
