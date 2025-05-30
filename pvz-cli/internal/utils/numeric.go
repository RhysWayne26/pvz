package utils

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"strings"
)

func ValidatePositiveInt(name string, val *int) error {
	if val != nil && *val <= 0 {
		return apperrors.Newf(
			apperrors.ValidationFailed,
			"parameter %s must be positive, got %d", name, *val,
		)
	}
	return nil
}

func ValidatePositiveFloat(name string, val float64) error {
	if val <= 0 {
		return apperrors.Newf(
			apperrors.ValidationFailed,
			"%s must be greater than 0, got %.2f", name, val,
		)
	}
	return nil
}

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
