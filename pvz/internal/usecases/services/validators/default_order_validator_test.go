package validators_test

import (
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/usecases/builders"
	"testing"
	"time"

	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/clock"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services/validators"
)

// TestDefaultOrderValidator_ValidateAccept tests the ValidateAccept function of DefaultOrderValidator for various scenarios.
func TestDefaultOrderValidator_ValidateAccept(t *testing.T) {
	clk := &clock.FakeClock{}
	v := validators.NewDefaultOrderValidator(clk)
	now := clk.Now()
	tests := []struct {
		name      string
		order     models.Order
		req       requests.AcceptOrderRequest
		expectErr bool
		wantCode  string
	}{
		{
			name:      "ok",
			order:     builders.NewOrderBuilder(clk).WithID(0).Build(),
			req:       requests.AcceptOrderRequest{ExpiresAt: now.Add(time.Hour)},
			expectErr: false,
		},
		{
			name:      "expired",
			order:     builders.NewOrderBuilder(clk).WithID(0).Build(),
			req:       requests.AcceptOrderRequest{ExpiresAt: now.Add(-time.Hour)},
			expectErr: true,
			wantCode:  string(apperrors.ValidationFailed),
		},
		{
			name:      "duplicate",
			order:     builders.NewOrderBuilder(clk).WithID(42).Build(),
			req:       requests.AcceptOrderRequest{ExpiresAt: now.Add(time.Hour)},
			expectErr: true,
			wantCode:  string(apperrors.OrderAlreadyExists),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := v.ValidateAccept(tt.order, tt.req)
			if tt.expectErr {
				require.Error(t, err)
				require.Equal(t, tt.wantCode, apperrors.CodeFromError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDefaultOrderValidator_ValidateIssue tests the ValidateIssue method of DefaultOrderValidator for various input scenarios.
func TestDefaultOrderValidator_ValidateIssue(t *testing.T) {
	clk := &clock.FakeClock{}
	v := validators.NewDefaultOrderValidator(clk)
	now := clk.Now()
	baseOrder := models.Order{
		OrderID:   1,
		UserID:    100,
		Status:    models.Accepted,
		ExpiresAt: now.Add(time.Hour),
	}
	tests := []struct {
		name      string
		order     models.Order
		req       requests.IssueOrdersRequest
		expectErr bool
		wantCode  string
	}{
		{
			name:      "no IDs",
			order:     baseOrder,
			req:       requests.IssueOrdersRequest{UserID: 100, OrderIDs: nil},
			expectErr: true,
			wantCode:  string(apperrors.ValidationFailed),
		},
		{
			name:      "wrong user",
			order:     baseOrder,
			req:       requests.IssueOrdersRequest{UserID: 200, OrderIDs: []uint64{1}},
			expectErr: true,
			wantCode:  string(apperrors.ValidationFailed),
		},
		{
			name: "bad status",
			order: builders.NewOrderBuilder(clk).
				WithID(1).
				WithUserID(100).
				WithStatus(models.Issued).
				WithExpiresAt(now.Add(time.Hour)).
				Build(),
			req:       requests.IssueOrdersRequest{UserID: 100, OrderIDs: []uint64{1}},
			expectErr: true,
			wantCode:  string(apperrors.ValidationFailed),
		},
		{
			name: "expired",
			order: builders.NewOrderBuilder(clk).
				WithID(1).
				WithUserID(100).
				WithStatus(models.Accepted).
				WithExpiresAt(now.Add(-time.Minute)).
				Build(),
			req:       requests.IssueOrdersRequest{UserID: 100, OrderIDs: []uint64{1}},
			expectErr: true,
			wantCode:  string(apperrors.StorageExpired)},
		{
			name:      "ok",
			order:     baseOrder,
			req:       requests.IssueOrdersRequest{UserID: 100, OrderIDs: []uint64{1}},
			expectErr: false,
			wantCode:  ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := v.ValidateIssue(tt.order, tt.req)
			if tt.expectErr {
				require.Error(t, err)
				require.Equal(t, tt.wantCode, apperrors.CodeFromError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDefaultOrderValidator_ValidateClientReturn tests the validation logic for client return requests.
func TestDefaultOrderValidator_ValidateClientReturn(t *testing.T) {
	clk := &clock.FakeClock{}
	v := validators.NewDefaultOrderValidator(clk)
	now := clk.Now()
	baseOrder := builders.NewOrderBuilder(clk).
		WithID(2).
		WithUserID(200).
		WithStatus(models.Issued).
		WithUpdatedStatusAt(now.Add(-constants.ReturnWindow / 2)).
		Build()
	tests := []struct {
		name      string
		order     models.Order
		req       requests.ClientReturnsRequest
		expectErr bool
		wantCode  string
	}{
		{
			name:      "no IDs",
			order:     baseOrder,
			req:       requests.ClientReturnsRequest{UserID: 200, OrderIDs: nil},
			expectErr: true,
			wantCode:  string(apperrors.ValidationFailed),
		},
		{
			name:      "wrong user",
			order:     baseOrder,
			req:       requests.ClientReturnsRequest{UserID: 300, OrderIDs: []uint64{2}},
			expectErr: true,
			wantCode:  string(apperrors.ValidationFailed),
		},
		{
			name: "bad status",
			order: builders.NewOrderBuilder(clk).
				WithID(2).
				WithUserID(200).
				WithStatus(models.Accepted).
				WithUpdatedStatusAt(now).
				Build(),
			req:       requests.ClientReturnsRequest{UserID: 200, OrderIDs: []uint64{2}},
			expectErr: true,
			wantCode:  string(apperrors.ValidationFailed),
		},
		{
			name: "window expired",
			order: builders.NewOrderBuilder(clk).
				WithID(2).
				WithUserID(200).
				WithStatus(models.Issued).
				WithUpdatedStatusAt(now.Add(-constants.ReturnWindow * 2)).
				Build(),
			req:       requests.ClientReturnsRequest{UserID: 200, OrderIDs: []uint64{2}},
			expectErr: true,
			wantCode:  string(apperrors.ValidationFailed),
		},
		{
			name:      "ok",
			order:     baseOrder,
			req:       requests.ClientReturnsRequest{UserID: 200, OrderIDs: []uint64{2}},
			expectErr: false,
			wantCode:  "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := v.ValidateClientReturn(tt.order, tt.req)
			if tt.expectErr {
				require.Error(t, err)
				require.Equal(t, tt.wantCode, apperrors.CodeFromError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultOrderValidator_ValidateReturnToCourier(t *testing.T) {
	clk := &clock.FakeClock{}
	v := validators.NewDefaultOrderValidator(clk)
	now := clk.Now()
	tests := []struct {
		name      string
		order     models.Order
		expectErr bool
		wantCode  string
	}{
		{
			name: "already returned",
			order: builders.
				NewOrderBuilder(clk).
				WithStatus(models.Returned).
				Build(),
			expectErr: false,
			wantCode:  "",
		},
		{
			name: "issued",
			order: builders.
				NewOrderBuilder(clk).
				WithID(3).
				WithStatus(models.Issued).
				Build(),
			expectErr: true,
			wantCode:  string(apperrors.OrderNotFound),
		},
		{
			name: "not expired",
			order: builders.
				NewOrderBuilder(clk).
				WithID(4).
				WithStatus(models.Accepted).
				WithExpiresAt(now.Add(time.Hour)).
				Build(),
			expectErr: true,
			wantCode:  string(apperrors.StorageExpired),
		},
		{
			name: "expired",
			order: builders.
				NewOrderBuilder(clk).
				WithID(5).
				WithStatus(models.Accepted).
				WithExpiresAt(now.Add(-time.Hour)).
				Build(),
			expectErr: false,
			wantCode:  "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := v.ValidateReturnToCourier(tt.order)
			if tt.expectErr {
				require.Error(t, err)
				require.Equal(t, tt.wantCode, apperrors.CodeFromError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
