package services

import "pvz-cli/internal/models"

// PackagePricingService calculates package pricing and validates weight constraints
type PackagePricingService interface {
	Evaluate(pkg models.PackageType, weight, price float64) (surcharge float64, err error)
}
