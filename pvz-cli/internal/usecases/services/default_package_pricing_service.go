package services

import (
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/models"
	"pvz-cli/internal/validators"
)

// DefaultPackagePricingService implements package pricing business logic with validation
type DefaultPackagePricingService struct {
	validator validators.PackageValidator
}

// NewDefaultPackagePricingService creates a new package pricing service with validator
func NewDefaultPackagePricingService(v validators.PackageValidator) *DefaultPackagePricingService {
	return &DefaultPackagePricingService{validator: v}
}

// Evaluate calculates package surcharge and validates weight constraints for given package type
func (s *DefaultPackagePricingService) Evaluate(pkg models.PackageType, weight float64, price float64) (float64, error) {
	if weight <= 0 {
		return 0, apperrors.Newf(apperrors.ValidationFailed, "weight must be > 0")
	}
	if price <= 0 {
		return 0, apperrors.Newf(apperrors.ValidationFailed, "price must be > 0")
	}

	if err := s.validator.Validate(pkg, weight); err != nil {
		return 0, err
	}

	var surcharge float64
	switch pkg {
	case models.PackageNone:
		surcharge = 0
	case models.PackageBag:
		surcharge = 5
	case models.PackageBox:
		surcharge = 20
	case models.PackageFilm:
		surcharge = 1
	case models.PackageBagFilm:
		surcharge = 5 + 1
	case models.PackageBoxFilm:
		surcharge = 20 + 1
	}

	totalPrice := price + surcharge
	return totalPrice, nil
}
