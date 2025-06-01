package validators

import (
	"pvz-cli/internal/constants"
	"time"

	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// DefaultOrderValidator is a default implementation of the OrderValidator interface.
type DefaultOrderValidator struct{}

// NewDefaultOrderValidator creates a new instance of DefaultOrderValidator.
func NewDefaultOrderValidator() *DefaultOrderValidator {
	return &DefaultOrderValidator{}
}

// ValidateAccept validates order acceptance requirements including expiry date and duplicates
func (v *DefaultOrderValidator) ValidateAccept(o models.Order, req requests.AcceptOrderRequest) error {
	if req.ExpiresAt.Before(time.Now()) {
		return apperrors.Newf(apperrors.ValidationFailed, "expires date is in the past")
	}
	if o.OrderID != "" {
		return apperrors.Newf(apperrors.OrderAlreadyExists, "order already exists")
	}
	return nil
}

// ValidateIssue validates order issuance requirements including user ownership and status
func (v *DefaultOrderValidator) ValidateIssue(orders []models.Order, req requests.IssueOrdersRequest) error {
	if len(req.OrderIDs) == 0 {
		return apperrors.Newf(apperrors.ValidationFailed, "no order IDs provided")
	}

	now := time.Now()
	for _, o := range orders {
		if o.UserID != req.UserID {
			return apperrors.Newf(apperrors.ValidationFailed, "order %s belongs to different user", o.OrderID)
		}
		if o.Status != models.Accepted {
			return apperrors.Newf(apperrors.ValidationFailed, "order %s status is %s, not ACCEPTED", o.OrderID, o.Status)
		}
		if o.ExpiresAt.Before(now) {
			return apperrors.Newf(apperrors.StorageExpired, "order %s storage period expired", o.OrderID)
		}
	}
	return nil
}

// ValidateClientReturn validates client return requests including ownership and return window
func (v *DefaultOrderValidator) ValidateClientReturn(orders []models.Order, req requests.ClientReturnsRequest) error {
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
func (v *DefaultOrderValidator) ValidateReturnToCourier(o models.Order) error {
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
