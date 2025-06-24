package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"strconv"
	"strings"
	"time"
)

// MapAcceptOrderParams converts CLI params for accept order command into internal request model
func (f *DefaultCLIFacadeMapper) MapAcceptOrderParams(p params.AcceptOrderParams) (requests.AcceptOrderRequest, error) {
	orderID, err := strconv.ParseUint(strings.TrimSpace(p.OrderID), 10, 64)
	if err != nil {
		return requests.AcceptOrderRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid order_id format")
	}

	userID, err := strconv.ParseUint(strings.TrimSpace(p.UserID), 10, 64)
	if err != nil {
		return requests.AcceptOrderRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid user_id format")
	}

	expiresAt, err := time.Parse(constants.TimeLayout, strings.TrimSpace(p.ExpiresAt))
	if err != nil {
		return requests.AcceptOrderRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid expires_at format")
	}

	weight, err := parseFloat("weight", p.Weight, constants.WeightFractionDigit)
	if err != nil {
		return requests.AcceptOrderRequest{}, err
	}

	price, err := parseFloat("price", p.Price, constants.PriceFractionDigit)
	if err != nil {
		return requests.AcceptOrderRequest{}, err
	}

	pkg, err := parsePackageType(p.Package)
	if err != nil {
		return requests.AcceptOrderRequest{}, err
	}

	return requests.AcceptOrderRequest{
		OrderID:   orderID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Weight:    weight,
		Price:     price,
		Package:   pkg,
	}, nil
}

func parseFloat(name, raw string, maxDigits int) (float32, error) {
	val64, err := strconv.ParseFloat(strings.TrimSpace(raw), 32)
	val := float32(val64)
	if err != nil {
		return 0, apperrors.Newf(apperrors.ValidationFailed, "invalid %s format", name)
	}
	if err := utils.ValidatePositiveFloat(name, val); err != nil {
		return 0, err
	}
	if err := utils.ValidateFractionDigits(name, val, maxDigits); err != nil {
		return 0, err
	}
	return val, nil
}

func parsePackageType(raw string) (models.PackageType, error) {
	normalized := strings.Trim(strings.TrimSpace(raw), `"`)
	if normalized == "" || strings.EqualFold(normalized, "null") {
		return models.PackageNone, nil
	}
	switch strings.ToLower(normalized) {
	case "none":
		return models.PackageNone, nil
	case "bag":
		return models.PackageBag, nil
	case "box":
		return models.PackageBox, nil
	case "film", "tape":
		return models.PackageFilm, nil
	case "bag+film", "bag_film", "bagfilm":
		return models.PackageBagFilm, nil
	case "box+film", "box_film", "boxfilm":
		return models.PackageBoxFilm, nil
	default:
		return models.PackageNone, apperrors.Newf(apperrors.ValidationFailed, "invalid package type: %s", normalized)
	}
}
