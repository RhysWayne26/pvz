package repositories

import "pvz-cli/internal/models"

type ReturnRepository interface {
	Save(ret models.ReturnEntry) error
	List(page, limit int) ([]models.ReturnEntry, error)
}
