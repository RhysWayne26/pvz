package validators

import (
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/models"
)

// DefaultPackageValidator is a default implementation of the PackageValidator interface.
type DefaultPackageValidator struct{}

// NewDefaultPackageValidator creates a new instance of DefaultPackageValidator
func NewDefaultPackageValidator() *DefaultPackageValidator {
	return &DefaultPackageValidator{}
}

// Validate checks if package type is valid and supports given weight
func (v *DefaultPackageValidator) Validate(pkg models.PackageType, weight float32) error {
	if !v.isValidPackageType(pkg) {
		return apperrors.Newf(apperrors.InvalidPackage, "package type is not valid")
	}
	switch pkg {
	case models.PackageBag, models.PackageBagFilm:
		if weight >= 10 {
			return apperrors.Newf(apperrors.WeightTooHeavy, "bag not suitable for weight >= 10kg")
		}
	case models.PackageBox, models.PackageBoxFilm:
		if weight >= 30 {
			return apperrors.Newf(apperrors.WeightTooHeavy, "box not suitable for weight >= 30kg")
		}
	}
	return nil
}

func (v *DefaultPackageValidator) isValidPackageType(pkg models.PackageType) bool {
	switch pkg {
	case models.PackageNone, models.PackageBag, models.PackageBox,
		models.PackageFilm, models.PackageBagFilm, models.PackageBoxFilm:
		return true
	default:
		return false
	}
}
