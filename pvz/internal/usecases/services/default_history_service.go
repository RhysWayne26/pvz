package services

import (
	"context"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

var _ HistoryService = (*DefaultHistoryService)(nil)

// DefaultHistoryService is a default implementation of the HistoryService interface
type DefaultHistoryService struct {
	historyRepo repositories.HistoryRepository
}

// NewDefaultHistoryService creates a new instance of DefaultHistoryService
func NewDefaultHistoryService(historyRepo repositories.HistoryRepository) *DefaultHistoryService {
	return &DefaultHistoryService{historyRepo: historyRepo}
}

// Record saves a history entry to the repository
func (s *DefaultHistoryService) Record(ctx context.Context, e models.HistoryEntry) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := s.historyRepo.Save(ctx, e); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to save history entry")
	}
	return nil
}

// List retrieves a list of history entries matching the specified filter, sorted by timestamp in ascending order.
func (s *DefaultHistoryService) List(ctx context.Context, filter requests.OrderHistoryFilter) ([]models.HistoryEntry, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	entries, count, err := s.historyRepo.List(ctx, filter)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to load history list: %v", err)
	}
	if count == 0 && filter.OrderID != nil {
		return nil, apperrors.Newf(apperrors.OrderNotFound, "order %d not found", *filter.OrderID)
	}
	return entries, nil
}
