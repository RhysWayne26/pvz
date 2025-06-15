package services

import (
	"context"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"sort"
)

// DefaultHistoryService is a default implementation of HistoryService interface
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

// GetByOrder retrieves all history entries for a specific order, sorted by timestamp
func (s *DefaultHistoryService) GetByOrder(ctx context.Context, orderID uint64) ([]models.HistoryEntry, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	entries, err := s.historyRepo.LoadByOrder(ctx, orderID)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to load history for order %d: %v", orderID, err)

	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	return entries, nil
}

// ListAll retrieves paginated list of all history entries, sorted by timestamp
func (s *DefaultHistoryService) ListAll(ctx context.Context, page, limit int) ([]models.HistoryEntry, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	entries, err := s.historyRepo.LoadAll(ctx, page, limit)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to load history list: %v", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	return entries, nil
}
