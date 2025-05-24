package services

import (
	"encoding/json"
	"fmt"
	"os"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/validators"
	"strings"
	"time"
)

type defaultOrderService struct {
	orderRepo  repositories.OrderRepository
	historySvc HistoryService
	validator  validators.OrderValidator
}

func NewDefaultOrderService(orderRepo repositories.OrderRepository, historyService HistoryService, validator validators.OrderValidator) OrderService {
	return &defaultOrderService{
		orderRepo:  orderRepo,
		historySvc: historyService,
		validator:  validator,
	}
}

func (s *defaultOrderService) AcceptOrder(req requests.AcceptOrderRequest) error {
	existing, err := s.orderRepo.Load(req.OrderID)
	if err != nil {
		existing = models.Order{}
	}

	if err := s.validator.ValidateAccept(existing, req); err != nil {
		return err
	}

	order := models.Order{
		OrderID:   req.OrderID,
		UserID:    req.UserID,
		Status:    models.Accepted,
		ExpiresAt: req.ExpiresAt,
	}
	if err := s.orderRepo.Save(order); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to save order: %v", err)
	}

	entry := models.HistoryEntry{
		OrderID:   order.OrderID,
		Event:     models.EventAccepted,
		Timestamp: time.Now(),
	}
	if err := s.historySvc.Record(entry); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to record history: %v", err)
	}

	return nil
}

func (s *defaultOrderService) IssueOrder(req requests.IssueOrderRequest) error {
	var orders []models.Order
	for _, id := range req.OrderIDs {
		o, err := s.orderRepo.Load(id)
		if err != nil {
			return apperrors.Newf(apperrors.OrderNotFound, "order %s not found", id)
		}
		orders = append(orders, o)
	}

	if err := s.validator.ValidateIssue(orders, req); err != nil {
		return err
	}

	now := time.Now()
	for _, o := range orders {
		o.Status = models.Issued
		o.IssuedAt = &now

		if err := s.orderRepo.Save(o); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to save order %s: %v", o.OrderID, err)
		}

		entry := models.HistoryEntry{
			OrderID:   o.OrderID,
			Event:     models.EventIssued,
			Timestamp: now,
		}
		if err := s.historySvc.Record(entry); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to record history for order %s: %v", o.OrderID, err)
		}
	}

	return nil
}

func (s *defaultOrderService) ListOrders(filter requests.ListOrdersFilter) ([]models.Order, string, int, error) {
	all, err := s.orderRepo.List(requests.ListOrdersFilter{
		UserID: filter.UserID,
		InPvz:  filter.InPvz,
	})
	if err != nil {
		return nil, "", 0, apperrors.Newf(apperrors.InternalError, "failed to list orders: %v", err)
	}

	if filter.Last != nil {
		n := *filter.Last
		if len(all) > n {
			all = all[len(all)-n:]
		}
		return all, "", len(all), nil
	}

	filtered, err := s.orderRepo.List(filter)
	if err != nil {
		return nil, "", 0, apperrors.Newf(apperrors.InternalError, "failed to list paginated: %v", err)
	}

	var nextLast string
	if len(filtered) > 0 {
		last := filtered[len(filtered)-1].OrderID
		nextLast = last
	}

	countFilter := filter
	countFilter.Page = nil
	countFilter.Limit = nil
	countFilter.LastID = ""

	counted, err := s.orderRepo.List(countFilter)
	if err != nil {
		return nil, "", 0, apperrors.Newf(apperrors.InternalError, "failed to count orders: %v", err)
	}

	return filtered, nextLast, len(counted), nil
}
func (s *defaultOrderService) ImportOrders(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, apperrors.Newf(apperrors.InternalError, "cannot open file %q: %v", filePath, err)
	}
	defer func() {
		err := f.Close()
		if err != nil {

		}
	}()

	var raw []map[string]string
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return 0, apperrors.Newf(apperrors.ValidationFailed, "invalid JSON: %v", err)
	}

	var imported int
	for _, item := range raw {
		orderID := strings.TrimSpace(item["order_id"])
		userID := strings.TrimSpace(item["user_id"])
		expiresAt, err := time.Parse(constants.TimeLayout, item["expires_at"])
		if err != nil {
			fmt.Printf("SKIPPED: %s — parsing failed\n", item["order_id"])
			continue
		}

		req := requests.AcceptOrderRequest{
			OrderID:   orderID,
			UserID:    userID,
			ExpiresAt: expiresAt,
		}

		if err := s.AcceptOrder(req); err != nil {
			fmt.Printf("ERROR importing %s: %v\n", item["order_id"], err)
		} else {
			imported++
		}
	}

	return imported, nil
}
