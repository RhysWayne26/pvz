package utils

import (
	"math"
	"pvz-cli/internal/common/apperrors"
)

// ValidatePositiveInt validates that integer parameter is positive
func ValidatePositiveInt(name string, val *int) error {
	if val != nil && *val <= 0 {
		return apperrors.Newf(
			apperrors.ValidationFailed,
			"parameter %s must be positive, got %d", name, *val,
		)
	}
	return nil
}

// ValidatePositiveFloat validates that float parameter is positive
func ValidatePositiveFloat(name string, val float32) error {
	if val <= 0 {
		return apperrors.Newf(
			apperrors.ValidationFailed,
			"%s must be greater than 0, got %.2f", name, val,
		)
	}
	return nil
}

// ValidateFractionDigits validates that float has no more than specified fractional digits
func ValidateFractionDigits(name string, value float32, maxDigits int) error {
	v := float64(value)
	mult := math.Pow10(maxDigits)
	vMult := v * mult
	intPart := math.Round(vMult)

	const epsilon = 1e-3
	if math.Abs(vMult-intPart) > epsilon {
		return apperrors.Newf(
			apperrors.ValidationFailed,
			"%s must have at most %d fractional digits", name, maxDigits,
		)
	}
	return nil
}
