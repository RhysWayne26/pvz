package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
	"strconv"
	"strings"
	"time"
)

// HandleAcceptOrderCommand processes accept-order command with package pricing validation, optionally suppressing output for batch import
func HandleAcceptOrderCommand(params dto.AcceptOrderParams, svc services.OrderService, silent bool) error {
	orderID := strings.TrimSpace(params.OrderID)
	userID := strings.TrimSpace(params.UserID)

	expiresAt, err := time.Parse(constants.TimeLayout, strings.TrimSpace(params.ExpiresAt))
	if err != nil {
		return apperrors.Newf(apperrors.ValidationFailed, "invalid expires_at format")
	}

	weight, err := handlePositiveFloatParam("weight", params.Weight, constants.WeightFractionDigit)
	if err != nil {
		return err
	}

	price, err := handlePositiveFloatParam("price", params.Price, constants.PriceFractionDigit)
	if err != nil {
		return err
	}

	pkg := normalizePackageType(params.Package)
	req := requests.AcceptOrderRequest{
		OrderID:   orderID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Weight:    weight,
		Price:     price,
		Package:   pkg,
	}

	order, err := svc.AcceptOrder(req)
	if err != nil {
		return err
	}

	if silent {
		return nil
	}
	fmt.Printf("ORDER_ACCEPTED: %s\n", order.OrderID)
	fmt.Printf("PACKAGE: %s\n", order.Package)
	fmt.Printf("TOTAL_PRICE: %.*f\n", constants.PriceFractionDigit, order.Price)
	return nil
}

func handlePositiveFloatParam(name string, raw string, maxFractionDigits int) (float64, error) {
	val, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, apperrors.Newf(apperrors.ValidationFailed, "invalid %s format", name)
	}
	if err := utils.ValidatePositiveFloat(name, val); err != nil {
		return 0, err
	}
	if err := utils.ValidateFractionDigits(name, val, maxFractionDigits); err != nil {
		return 0, err
	}
	return val, nil
}

func normalizePackageType(raw string) models.PackageType {
	trimmed := strings.TrimSpace(raw)
	normalized := strings.Trim(trimmed, `"`)
	if normalized == "" || strings.EqualFold(normalized, "null") {
		return models.PackageNone
	}
	return models.PackageType(normalized)
}
