package services

import "pvz-cli/internal/models"

type PackagePricingService interface {
	Evaluate(pkg models.PackageType, weight, price float64) (surcharge float64, err error)
}
