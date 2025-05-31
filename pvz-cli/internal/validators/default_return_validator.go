package validators

import (
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"time"
)

// DefaultReturnValidator is a default implementation of the ReturnValidator interface.
type DefaultReturnValidator struct{}

// NewDefaultReturnValidator creates a new instance of DefaultReturnValidator
func NewDefaultReturnValidator() *DefaultReturnValidator {
	return &DefaultReturnValidator{}
}

// ValidateClientReturn validates client return requests including ownership and return window
func (v *DefaultReturnValidator) ValidateClientReturn(orders []models.Order, req requests.ClientReturnsRequest) error {
	if len(req.OrderIDs) == 0 {
		return apperrors.Newf(apperrors.ValidationFailed, "no order IDs provided")
	}

	now := time.Now()
	for _, o := range orders {
		if o.UserID != req.UserID {
			return apperrors.Newf(apperrors.ValidationFailed, "order %s belongs to another user", o.OrderID)
		}
		if o.Status != models.Issued {
			return apperrors.Newf(apperrors.ValidationFailed, "order %s status is %s, not ISSUED", o.OrderID, o.Status)
		}
		if o.IssuedAt == nil {
			return apperrors.Newf(apperrors.InternalError, "order %s missing issued_at", o.OrderID)
		}
		if now.Sub(*o.IssuedAt) > constants.ReturnWindow {
			return apperrors.Newf(apperrors.ValidationFailed, "return window expired for order %s", o.OrderID)
		}
	}
	return nil
}

// ValidateReturnToCourier validates order return to courier including status and expiration
func (v *DefaultReturnValidator) ValidateReturnToCourier(o models.Order) error {
	if o.Status == models.Returned {
		return nil
	}
	now := time.Now()
	if o.Status == models.Issued {
		return apperrors.Newf(apperrors.OrderNotFound, "order %s not found", o.OrderID)
	}
	if o.ExpiresAt.After(now) {
		return apperrors.Newf(apperrors.StorageExpired, "cannot return order %s before expiration", o.OrderID)
	}
	return nil
}
