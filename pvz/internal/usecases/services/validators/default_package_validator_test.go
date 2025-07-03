package validators

import (
	"github.com/stretchr/testify/require"
	"testing"

	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/models"
)

// TestDefaultPackageValidator_Validate tests the validation logic of DefaultPackageValidator for various package types and weights.
func TestDefaultPackageValidator_Validate(t *testing.T) {
	v := NewDefaultPackageValidator()
	tests := []struct {
		name      string
		pkg       models.PackageType
		weight    float32
		expectErr bool
		wantCode  string
	}{
		{
			name:      "valid: PackageNone",
			pkg:       models.PackageNone,
			weight:    0,
			expectErr: false,
		},
		{
			name:      "valid: PackageBag under 10kg",
			pkg:       models.PackageBag,
			weight:    9.999,
			expectErr: false,
		},
		{
			name:      "invalid: PackageBag overweight",
			pkg:       models.PackageBag,
			weight:    10,
			expectErr: true,
			wantCode:  string(apperrors.WeightTooHeavy),
		},
		{
			name:      "invalid: PackageBox overweight",
			pkg:       models.PackageBox,
			weight:    30,
			expectErr: true,
			wantCode:  string(apperrors.WeightTooHeavy),
		},
		{
			name:      "valid: PackageBox under 30kg",
			pkg:       models.PackageBox,
			weight:    29.999,
			expectErr: false,
		},
		{
			name:      "valid: PackageFilm no restriction",
			pkg:       models.PackageFilm,
			weight:    239,
			expectErr: false,
		},
		{
			name:      "invalid: Unknown package type",
			pkg:       models.PackageType(999),
			weight:    5,
			expectErr: true,
			wantCode:  string(apperrors.InvalidPackage),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := v.Validate(tt.pkg, tt.weight)
			if tt.expectErr {
				require.Error(t, err)
				require.Equal(t, tt.wantCode, apperrors.CodeFromError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
