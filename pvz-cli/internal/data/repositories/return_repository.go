package repositories

import "pvz-cli/internal/models"

// ReturnRepository handles persistence operations for return entries
type ReturnRepository interface {
	Save(ret models.ReturnEntry) error
	List(page, limit int) ([]models.ReturnEntry, error)
}
