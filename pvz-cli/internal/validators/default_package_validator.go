package validators

import (
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/models"
)

type defaultPackageValidator struct{}

// NewDefaultPackageValidator creates a new package validator
func NewDefaultPackageValidator() PackageValidator {
	return &defaultPackageValidator{}
}

// Validate checks if package type is valid and supports given weight
func (v *defaultPackageValidator) Validate(pkg models.PackageType, weight float64) error {
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

func (v *defaultPackageValidator) isValidPackageType(pkg models.PackageType) bool {
	switch pkg {
	case models.PackageNone, models.PackageBag, models.PackageBox,
		models.PackageFilm, models.PackageBagFilm, models.PackageBoxFilm:
		return true
	default:
		return false
	}
}
