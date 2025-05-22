package data

import (
	"pvz-cli/internal/models"
)

type Snapshot struct {
	Orders  []models.Order
	Returns []models.ReturnEntry
	History []models.HistoryEntry
}
