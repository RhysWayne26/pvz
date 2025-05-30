package validators

import (
	"time"

	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

type defaultOrderValidator struct{}

// NewDefaultOrderValidator creates a new order validator
func NewDefaultOrderValidator() OrderValidator {
	return &defaultOrderValidator{}
}

// ValidateAccept validates order acceptance requirements including expiry date and duplicates
func (v *defaultOrderValidator) ValidateAccept(o models.Order, req requests.AcceptOrderRequest) error {
	if req.ExpiresAt.Before(time.Now()) {
		return apperrors.Newf(apperrors.ValidationFailed, "expires date is in the past")
	}
	if o.OrderID != "" {
		return apperrors.Newf(apperrors.OrderAlreadyExists, "order already exists")
	}
	return nil
}

// ValidateIssue validates order issuance requirements including user ownership and status
func (v *defaultOrderValidator) ValidateIssue(orders []models.Order, req requests.IssueOrdersRequest) error {
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
