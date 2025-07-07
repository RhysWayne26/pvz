package services

import (
	"context"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services/validators"
	"pvz-cli/internal/workerpool"
	"pvz-cli/pkg/clock"
	"sync"
)

var _ OrderService = (*DefaultOrderService)(nil)

// DefaultOrderService is a default implementation of the OrderService interface
type DefaultOrderService struct {
	clk               clock.Clock
	pool              workerpool.WorkerPool
	orderRepo         repositories.OrderRepository
	packagePricingSvc PackagePricingService
	historySvc        HistoryService
	validator         validators.OrderValidator
}

// NewDefaultOrderService creates a new instance of DefaultOrderService
func NewDefaultOrderService(
	clk clock.Clock,
	pool workerpool.WorkerPool,
	orderRepo repositories.OrderRepository,
	packagePricingService PackagePricingService,
	historyService HistoryService,
	validator validators.OrderValidator) *DefaultOrderService {
	return &DefaultOrderService{
		clk:               clk,
		pool:              pool,
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

	now := s.clk.Now()

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
func (s *DefaultOrderService) IssueOrders(
	ctx context.Context,
	req requests.IssueOrdersRequest,
) ([]models.BatchEntryProcessedResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := len(req.OrderIDs)
	results := make([]models.BatchEntryProcessedResult, n)
	var wg sync.WaitGroup
	for i, id := range req.OrderIDs {
		wg.Add(1)
		i, id := i, id
		s.pool.Submit(func() {
			defer wg.Done()
			res := models.BatchEntryProcessedResult{OrderID: id}
			order, err := s.orderRepo.Load(ctx, id)
			if err != nil {
				res.Error = apperrors.Newf(apperrors.OrderNotFound, "order %d not found", id)
				results[i] = res
				return
			}
			if err := s.validator.ValidateIssue(order, req); err != nil {
				res.Error = err
				results[i] = res
				return
			}
			order.Status = models.Issued
			order.UpdatedStatusAt = s.clk.Now()
			if err := s.orderRepo.Save(ctx, order); err != nil {
				res.Error = apperrors.Newf(apperrors.InternalError, "failed to save order %d: %v", id, err)
				results[i] = res
				return
			}
			_ = s.historySvc.Record(ctx, models.HistoryEntry{
				OrderID:   id,
				Event:     models.EventIssued,
				Timestamp: s.clk.Now(),
			})

			results[i] = res
		})
	}

	wg.Wait()
	return results, nil
}

// ListOrders retrieves filtered and paginated list of orders
func (s *DefaultOrderService) ListOrders(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, uint64, int, error) {
	if ctx.Err() != nil {
		return nil, 0, 0, ctx.Err()
	}
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
func (s *DefaultOrderService) CreateClientReturns(
	ctx context.Context,
	req requests.ClientReturnsRequest,
) ([]models.BatchEntryProcessedResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := len(req.OrderIDs)
	results := make([]models.BatchEntryProcessedResult, n)
	var wg sync.WaitGroup
	for i, id := range req.OrderIDs {
		wg.Add(1)
		i, id := i, id
		s.pool.Submit(func() {
			defer wg.Done()
			res := models.BatchEntryProcessedResult{OrderID: id}
			order, err := s.orderRepo.Load(ctx, id)
			if err != nil {
				res.Error = apperrors.Newf(apperrors.OrderNotFound, "order %d not found", id)
				results[i] = res
				return
			}
			if err := s.validator.ValidateClientReturn(order, req); err != nil {
				res.Error = err
				results[i] = res
				return
			}
			order.Status = models.Returned
			order.UpdatedStatusAt = s.clk.Now()
			if err := s.orderRepo.Save(ctx, order); err != nil {
				res.Error = apperrors.Newf(apperrors.InternalError, "failed to save order %d: %v", id, err)
				results[i] = res
				return
			}

			_ = s.historySvc.Record(ctx, models.HistoryEntry{
				OrderID:   id,
				Event:     models.EventReturnedByClient,
				Timestamp: s.clk.Now(),
			})
			results[i] = res
		})
	}

	wg.Wait()
	return results, nil
}

// ReturnToCourier processes return of order back to courier/warehouse
func (s *DefaultOrderService) ReturnToCourier(ctx context.Context, req requests.ReturnOrderRequest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
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
		Timestamp: s.clk.Now(),
	}
	if err := s.historySvc.Record(ctx, entry); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to record history for order %d: %v", orderID, err)
	}

	return nil
}

// ListReturns retrieves paginated list of return entries sorted by return date
func (s *DefaultOrderService) ListReturns(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	orders, _, err := s.orderRepo.List(ctx, filter)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to list returns: %v", err)
	}

	return orders, nil
}

// ImportOrders imports multiple orders concurrently, processing each status and returning a batch of results with errors, if any.
func (s *DefaultOrderService) ImportOrders(
	ctx context.Context,
	req requests.ImportOrdersRequest,
) ([]models.BatchEntryProcessedResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := len(req.Statuses)
	results := make([]models.BatchEntryProcessedResult, n)
	var wg sync.WaitGroup
	for i, st := range req.Statuses {
		if st.Error != nil {
			results[i] = models.BatchEntryProcessedResult{OrderID: st.Request.OrderID, Error: st.Error}
			continue
		}
		wg.Add(1)
		i, st := i, st
		s.pool.Submit(func() {
			defer wg.Done()
			order, err := s.AcceptOrder(ctx, *st.Request)
			results[i] = models.BatchEntryProcessedResult{OrderID: order.OrderID, Error: err}
		})
	}
	wg.Wait()
	return results, nil
}
