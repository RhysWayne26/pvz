package strategies

import "pvz-cli/internal/models"

type defaultPricingStrategy struct{}

// NewDefaultPricingStrategy creates a new instance of the default pricing strategy.
func NewDefaultPricingStrategy() PricingStrategy {
	return &defaultPricingStrategy{}
}

// GetSurcharge returns the surcharge amount for the given package type according to the default pricing rules.
func (d *defaultPricingStrategy) GetSurcharge(pkg models.PackageType) float64 {
	switch pkg {
	case models.PackageNone:
		return 0
	case models.PackageBag:
		return 5
	case models.PackageBox:
		return 20
	case models.PackageFilm:
		return 1
	case models.PackageBagFilm:
		return 6
	case models.PackageBoxFilm:
		return 21
	default:
		return 0
	}
}
