package utils

import "pvz-cli/internal/apperrors"

func ValidatePositiveInt(name string, val *int) error {
	if val != nil && *val <= 0 {
		return apperrors.Newf(
			apperrors.ValidationFailed,
			"parameter %s must be positive, got %d", name, *val,
		)
	}
	return nil
}
