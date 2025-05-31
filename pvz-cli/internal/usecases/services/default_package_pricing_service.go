package services

import (
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/strategies"
	"pvz-cli/internal/validators"
)

// DefaultPackagePricingService implements package pricing business logic with validation
type DefaultPackagePricingService struct {
	validator validators.PackageValidator
	strategy  strategies.PricingStrategy
}

// NewDefaultPackagePricingService creates a new package pricing service with validator
func NewDefaultPackagePricingService(v validators.PackageValidator, s strategies.PricingStrategy) *DefaultPackagePricingService {
	return &DefaultPackagePricingService{
		validator: v,
		strategy:  s,
	}
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

	surcharge := s.strategy.GetSurcharge(pkg)
	return price + surcharge, nil
}
