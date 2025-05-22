package repositories

import (
	"github.com/google/uuid"
	"pvz-cli/internal/models"
)

type HistoryRepository interface {
	Save(e models.HistoryEntry) error
	LoadByOrder(orderID uuid.UUID) ([]models.HistoryEntry, error)
	LoadAll(page, limit int) ([]models.HistoryEntry, error)
}
