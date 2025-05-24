package repositories

import (
	"pvz-cli/internal/models"
)

type HistoryRepository interface {
	Save(e models.HistoryEntry) error
	LoadByOrder(orderID string) ([]models.HistoryEntry, error)
	LoadAll(page, limit int) ([]models.HistoryEntry, error)
}
