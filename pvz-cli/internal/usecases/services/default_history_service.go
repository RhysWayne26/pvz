package services

import (
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"sort"
)

type defaultHistoryService struct {
	historyRepo repositories.HistoryRepository
}

func NewDefaultHistoryService(historyRepo repositories.HistoryRepository) HistoryService {
	return &defaultHistoryService{historyRepo: historyRepo}
}

func (s *defaultHistoryService) Record(e models.HistoryEntry) error {
	if err := s.historyRepo.Save(e); err != nil {
		return apperrors.Newf(apperrors.InternalError, "failed to save history entry")
	}
	return nil
}

func (s *defaultHistoryService) GetByOrder(orderID string) ([]models.HistoryEntry, error) {
	entries, err := s.historyRepo.LoadByOrder(orderID)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to load history for order %s: %v", orderID, err)

	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	return entries, nil
}

func (s *defaultHistoryService) ListAll(page, limit int) ([]models.HistoryEntry, error) {
	entries, err := s.historyRepo.LoadAll(page, limit)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to load history list: %v", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	return entries, nil
}
