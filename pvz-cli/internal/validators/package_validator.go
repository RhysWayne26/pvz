package validators

import "pvz-cli/internal/models"

type PackageValidator interface {
	Validate(pkg models.PackageType, weight float64) error
}
