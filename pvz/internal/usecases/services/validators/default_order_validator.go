package validators

import (
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/pkg/clock"
)

var _ OrderValidator = (*DefaultOrderValidator)(nil)

// DefaultOrderValidator is a default implementation of the OrderValidator interface.
type DefaultOrderValidator struct {
	clk clock.Clock
}

// NewDefaultOrderValidator creates a new instance of DefaultOrderValidator.
func NewDefaultOrderValidator(clk clock.Clock) *DefaultOrderValidator {
	return &DefaultOrderValidator{
		clk: clk,
	}
}

// ValidateAccept validates order acceptance requirements including expiry date and duplicates
func (v *DefaultOrderValidator) ValidateAccept(o models.Order, req requests.AcceptOrderRequest) error {
	if req.ExpiresAt.Before(v.clk.Now()) {
		return apperrors.Newf(apperrors.ValidationFailed, "expires date is in the past")
	}
	if o.OrderID != 0 {
		return apperrors.Newf(apperrors.OrderAlreadyExists, "order already exists")
	}
	return nil
}

// ValidateIssue validates order issuance requirements including user ownership and status
func (v *DefaultOrderValidator) ValidateIssue(o models.Order, req requests.IssueOrdersRequest) error {
	if len(req.OrderIDs) == 0 {
		return apperrors.Newf(apperrors.ValidationFailed, "no order IDs provided")
	}
	now := v.clk.Now()
	if o.UserID != req.UserID {
		return apperrors.Newf(apperrors.ValidationFailed, "order %d belongs to different user", o.OrderID)
	}
	if o.Status != models.Accepted {
		return apperrors.Newf(apperrors.ValidationFailed, "order %d status is %s, not ACCEPTED", o.OrderID, o.Status)
	}
	if o.ExpiresAt.Before(now) {
		return apperrors.Newf(apperrors.StorageExpired, "order %d storage period expired", o.OrderID)
	}
	return nil
}

// ValidateClientReturn validates client return requests including ownership and return window
func (v *DefaultOrderValidator) ValidateClientReturn(o models.Order, req requests.ClientReturnsRequest) error {
	if len(req.OrderIDs) == 0 {
		return apperrors.Newf(apperrors.ValidationFailed, "no order IDs provided")
	}
	now := v.clk.Now()
	if o.UserID != req.UserID {
		return apperrors.Newf(apperrors.ValidationFailed, "order %d belongs to another user", o.OrderID)
	}
	if o.Status != models.Issued {
		return apperrors.Newf(apperrors.ValidationFailed, "order %d status is %s, not ISSUED", o.OrderID, o.Status)
	}

	if now.Sub(o.UpdatedStatusAt) > constants.ReturnWindow {
		return apperrors.Newf(apperrors.ValidationFailed, "return window expired for order %d", o.OrderID)
	}
	return nil
}

// ValidateReturnToCourier validates order return to courier including status and expiration
func (v *DefaultOrderValidator) ValidateReturnToCourier(o models.Order) error {
	if o.Status == models.Returned {
		return nil
	}
	now := v.clk.Now()
	if o.Status == models.Issued {
		return apperrors.Newf(apperrors.OrderNotFound, "order %d not found", o.OrderID)
	}
	if o.ExpiresAt.After(now) {
		return apperrors.Newf(apperrors.StorageExpired, "cannot return order %d before expiration", o.OrderID)
	}
	return nil
}
