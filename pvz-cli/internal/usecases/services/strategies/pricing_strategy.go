package strategies

import "pvz-cli/internal/models"

// PricingStrategy defines the interface for calculating a surcharge based on the selected package type.
type PricingStrategy interface {
	GetSurcharge(pkg models.PackageType) float64
}
