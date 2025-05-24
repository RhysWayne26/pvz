package services

import (
	"pvz-cli/internal/models"
)

type HistoryService interface {
	Record(e models.HistoryEntry) error
	GetByOrder(orderID string) ([]models.HistoryEntry, error)
	ListAll(page, limit int) ([]models.HistoryEntry, error)
}
