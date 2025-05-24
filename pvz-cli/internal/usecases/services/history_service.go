package services

import (
	"github.com/google/uuid"
	"pvz-cli/internal/models"
)

type HistoryService interface {
	Record(e models.HistoryEntry) error
	GetByOrder(orderID uuid.UUID) ([]models.HistoryEntry, error)
	ListAll(page, limit int) ([]models.HistoryEntry, error)
}
