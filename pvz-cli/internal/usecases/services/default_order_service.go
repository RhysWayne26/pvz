package services

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/common"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/validators"
	"time"
)

type defaultOrderService struct {
	orderRepo         repositories.OrderRepository
	packagePricingSvc PackagePricingService
	historySvc        HistoryService
	validator         validators.OrderValidator
}

// NewDefaultOrderService creates a new order service with all required dependencies
func NewDefaultOrderService(
	orderRepo repositories.OrderRepository, packagePricingService PackagePricingService, historyService HistoryService, validator validators.OrderValidator) OrderService {
	return &defaultOrderService{
		orderRepo:         orderRepo,
		packagePricingSvc: packagePricingService,
		historySvc:        historyService,
		validator:         validator,
	}
}

// AcceptOrder accepts an order with package pricing calculation and validation
func (s *defaultOrderService) AcceptOrder(req requests.AcceptOrderRequest) (models.Order, error) {
	existing, err := s.orderRepo.Load(req.OrderID)
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

	order := models.Order{
		OrderID:   req.OrderID,
		UserID:    req.UserID,
		CreatedAt: time.Now(),
		Status:    models.Accepted,
		ExpiresAt: req.ExpiresAt,
		Weight:    req.Weight,
		Price:     totalPrice,
		Package:   req.Package,
	}
	if err := s.orderRepo.Save(order); err != nil {
		return models.Order{}, apperrors.Newf(apperrors.InternalError, "failed to save order: %v", err)
	}

	entry := models.HistoryEntry{
		OrderID:   order.OrderID,
		Event:     models.EventAccepted,
		Timestamp: time.Now(),
	}
	if err := s.historySvc.Record(entry); err != nil {
		return models.Order{}, apperrors.Newf(apperrors.InternalError, "failed to record history: %v", err)
	}

	return order, nil
}

// IssueOrders processes multiple orders for issuance to clients
func (s *defaultOrderService) IssueOrders(req requests.IssueOrdersRequest) []common.ProcessResult {
	results := make([]common.ProcessResult, 0, len(req.OrderIDs))
	now := time.Now()

	for _, id := range req.OrderIDs {
		res := common.ProcessResult{OrderID: id}

		order, err := s.orderRepo.Load(id)
		if err != nil {
			res.Error = apperrors.Newf(apperrors.OrderNotFound, "order %s not found", id)
			results = append(results, res)
			continue
		}

		if err := s.validator.ValidateIssue([]models.Order{order}, req); err != nil {
			res.Error = err
			results = append(results, res)
			continue
		}

		order.Status = models.Issued
		order.IssuedAt = &now

		if err := s.orderRepo.Save(order); err != nil {
			res.Error = apperrors.Newf(apperrors.InternalError, "failed to save order %s: %v", id, err)
			results = append(results, res)
			continue
		}

		entry := models.HistoryEntry{
			OrderID:   order.OrderID,
			Event:     models.EventIssued,
			Timestamp: now,
		}

		if err := s.historySvc.Record(entry); err != nil {
			fmt.Printf("WARNING: failed to record history for order %s: %v\n", id, err)
		}

		res.Error = nil
		results = append(results, res)
	}

	return results
}

// ListOrders retrieves filtered and paginated list of orders
func (s *defaultOrderService) ListOrders(filter requests.ListOrdersFilter) ([]models.Order, string, int, error) {
	result, total, err := s.orderRepo.List(filter)
	if err != nil {
		return nil, "", 0, apperrors.Newf(apperrors.InternalError, "failed to list orders: %v", err)
	}

	if filter.Last != nil {
		n := *filter.Last
		if len(result) > n {
			result = result[len(result)-n:]
		}
		total = len(result)
	}

	var nextLastID string
	if len(result) > 0 {
		nextLastID = result[len(result)-1].OrderID
	}

	return result, nextLastID, total, nil
}
