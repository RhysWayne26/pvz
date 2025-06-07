package utils

import (
	"fmt"
	"pvz-cli/internal/common/apperrors"
	"strings"
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
func ValidatePositiveFloat(name string, val float64) error {
	if val <= 0 {
		return apperrors.Newf(
			apperrors.ValidationFailed,
			"%s must be greater than 0, got %.2f", name, val,
		)
	}
	return nil
}

// ValidateFractionDigits validates that float has no more than specified fractional digits
func ValidateFractionDigits(name string, value float64, maxDigits int) error {
	s := fmt.Sprintf("%.10f", value)
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return nil
	}
	frac := strings.TrimRight(parts[1], "0")
	if len(frac) > maxDigits {
		return apperrors.Newf(
			apperrors.ValidationFailed,
			"%s must have at most %d fractional digits", name, maxDigits,
		)
	}
	return nil
}
