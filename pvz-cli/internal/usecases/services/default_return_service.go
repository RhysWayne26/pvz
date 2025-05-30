package services

import (
	"fmt"
	"pvz-cli/internal/usecases/common"
	"sort"
	"time"

	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/validators"
)

type defaultReturnService struct {
	orderRepo  repositories.OrderRepository
	returnRepo repositories.ReturnRepository
	historySvc HistoryService
	validator  validators.ReturnValidator
}

// NewDefaultReturnService creates a new return service with all required dependencies
func NewDefaultReturnService(
	orderRepo repositories.OrderRepository,
	returnRepo repositories.ReturnRepository,
	historySvc HistoryService,
	validator validators.ReturnValidator,
) ReturnService {
	return &defaultReturnService{
		orderRepo:  orderRepo,
		returnRepo: returnRepo,
		historySvc: historySvc,
		validator:  validator,
	}
}

// CreateClientReturns processes multiple client return requests
func (s *defaultReturnService) CreateClientReturns(req requests.ClientReturnsRequest) []common.ProcessResult {
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

		if err := s.validator.ValidateClientReturn([]models.Order{order}, req); err != nil {
			res.Error = err
			results = append(results, res)
			continue
		}

		order.Status = models.Returned
		order.ReturnedAt = &now

		if err := s.orderRepo.Save(order); err != nil {
			res.Error = apperrors.Newf(apperrors.InternalError, "failed to save order %s: %v", order.OrderID, err)
			results = append(results, res)
			continue
		}

		ret := models.ReturnEntry{
			OrderID:    order.OrderID,
			UserID:     order.UserID,
			ReturnedAt: now,
		}
		if err := s.returnRepo.Save(ret); err != nil {
			res.Error = apperrors.Newf(apperrors.InternalError, "failed to save return entry for order %s: %v", order.OrderID, err)
			results = append(results, res)
			continue
		}

		entry := models.HistoryEntry{
			OrderID:   order.OrderID,
			Event:     models.EventReturnedFromClient,
			Timestamp: now,
		}
		if err := s.historySvc.Record(entry); err != nil {
			fmt.Printf("WARNING: failed to record history for order %s: %v\n", order.OrderID, err)
		}

		res.Error = nil
		results = append(results, res)
	}

	return results
}

// ReturnToCourier processes return of order back to courier/warehouse
func (s *defaultReturnService) ReturnToCourier(req requests.ReturnOrderRequest) error {
	orderID := req.OrderID
	o, err := s.orderRepo.Load(orderID)
	if err != nil {
		return apperrors.Newf(apperrors.OrderNotFound, "order %s not found", orderID)
	}

	if err := s.validator.ValidateReturnToCourier(o); err != nil {
		return err
	}

	if err := s.orderRepo.Delete(orderID); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to delete order %s: %v", orderID, err)
	}

	entry := models.HistoryEntry{
		OrderID:   orderID,
		Event:     models.EventReturnedToWarehouse,
		Timestamp: time.Now(),
	}
	if err := s.historySvc.Record(entry); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to record history for order %s: %v", orderID, err)
	}

	return nil
}

// ListReturns retrieves paginated list of return entries sorted by return date
func (s *defaultReturnService) ListReturns(page, limit int) ([]models.ReturnEntry, error) {
	rets, err := s.returnRepo.List(page, limit)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to list returns: %v", err)
	}

	entries := make([]models.ReturnEntry, len(rets))
	for i, r := range rets {
		entries[i] = models.ReturnEntry{
			OrderID:    r.OrderID,
			UserID:     r.UserID,
			ReturnedAt: r.ReturnedAt,
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ReturnedAt.Before(entries[j].ReturnedAt)
	})

	return entries, nil
}
