package data

import (
	"pvz-cli/internal/models"
)

// Snapshot represents complete application state for persistence
type Snapshot struct {
	Orders  []models.Order
	Returns []models.ReturnEntry
	History []models.HistoryEntry
}
