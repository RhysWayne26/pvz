package services

import (
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

func (s *defaultReturnService) CreateClientReturn(req requests.ClientReturnRequest) error {
	orders := make([]models.Order, 0, len(req.OrderIDs))
	for _, id := range req.OrderIDs {
		o, err := s.orderRepo.Load(id)
		if err != nil {
			return apperrors.Newf(apperrors.OrderNotFound, "order %s not found", id)
		}
		orders = append(orders, o)
	}

	if err := s.validator.ValidateClientReturn(orders, req); err != nil {
		return err
	}

	now := time.Now()
	for _, o := range orders {
		o.Status = models.Returned
		o.ReturnedAt = &now

		if err := s.orderRepo.Save(o); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to save order %s: %v", o.OrderID, err)
		}

		ret := models.ReturnEntry{
			OrderID:    o.OrderID,
			UserID:     o.UserID,
			ReturnedAt: now,
		}
		if err := s.returnRepo.Save(ret); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to save return entry for order %s: %v", o.OrderID, err)
		}

		entry := models.HistoryEntry{
			OrderID:   o.OrderID,
			Event:     models.EventReturnedFromClient,
			Timestamp: now,
		}
		if err := s.historySvc.Record(entry); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to record history for order %s: %v", o.OrderID, err)
		}
	}

	return nil
}

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
