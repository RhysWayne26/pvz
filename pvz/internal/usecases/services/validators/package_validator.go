package validators

import "pvz-cli/internal/models"

// PackageValidator validates package types and weight constraints
type PackageValidator interface {
	Validate(pkg models.PackageType, weight float32) error
}
