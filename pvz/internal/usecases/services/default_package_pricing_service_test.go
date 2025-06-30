package services

import (
	"pvz-cli/internal/usecases/services/strategies/mocks"
	valmocks "pvz-cli/internal/usecases/services/validators/mocks"
	"testing"

	"github.com/stretchr/testify/require"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/models"
)

// TestDefaultPackagePricingService_Evaluate_Success tests that the Evaluate method calculates the correct total price.
func TestDefaultPackagePricingService_Evaluate_Success(t *testing.T) {
	t.Parallel()
	v := valmocks.NewPackageValidatorMock(t)
	s := mocks.NewPricingStrategyMock(t)
	svc := NewDefaultPackagePricingService(v, s)
	pkg := models.PackageBox
	weight, price := float32(2), float32(100)
	surcharge := float32(25)
	v.ValidateMock.Expect(pkg, weight).Return(nil)
	s.GetSurchargeMock.Expect(pkg).Return(surcharge)
	got, err := svc.Evaluate(pkg, weight, price)
	require.NoError(t, err)
	require.Equal(t, price+surcharge, got)
}

// TestDefaultPackagePricingService_Evaluate_ValidationError verifies that a validation error is returned when validation fails.
func TestDefaultPackagePricingService_Evaluate_ValidationError(t *testing.T) {
	t.Parallel()
	v := valmocks.NewPackageValidatorMock(t)
	s := mocks.NewPricingStrategyMock(t)
	svc := NewDefaultPackagePricingService(v, s)
	pkg := models.PackageBox
	weight, price := float32(100), float32(100)
	vErr := apperrors.Newf(apperrors.ValidationFailed, "too heavy")
	v.ValidateMock.Expect(pkg, weight).Return(vErr)
	_, err := svc.Evaluate(pkg, weight, price)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	require.Equal(t, apperrors.ValidationFailed, ae.Code)
}

// TestDefaultPackagePricingService_Evaluate_InvalidParams verifies Evaluate method handles invalid inputs correctly.
func TestDefaultPackagePricingService_Evaluate_InvalidParams(t *testing.T) {
	t.Parallel()
	svc := NewDefaultPackagePricingService(nil, nil)
	cases := []struct {
		name   string
		weight float32
		price  float32
		want   apperrors.ErrorCode
	}{
		{"zero weight", 0, 100, apperrors.ValidationFailed},
		{"neg weight", -1, 100, apperrors.ValidationFailed},
		{"zero price", 10, 0, apperrors.ValidationFailed},
		{"neg price", 10, -1, apperrors.ValidationFailed},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := svc.Evaluate(models.PackageBox, tc.weight, tc.price)
			require.Error(t, err)
			var ae *apperrors.AppError
			require.ErrorAs(t, err, &ae)
			require.Equal(t, tc.want, ae.Code)
		})
	}
}
