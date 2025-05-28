package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/common"
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
		CreatedAt: time.Now(),
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
func (s *defaultOrderService) ImportOrders(filePath string) (int, error) {
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return 0, fmt.Errorf("path traversal not allowed")
	}
	f, err := os.Open(filePath)
	if err != nil {
		return 0, apperrors.Newf(apperrors.InternalError, "cannot open file %q: %v", filePath, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("WARNING: failed to close file: %v\n", err)
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
