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

// AcceptOrderHandler handles the accept order command.
type AcceptOrderHandler struct {
	params  dto.AcceptOrderParams
	service services.OrderService
	silent  bool
}

// NewAcceptOrderHandler creates an instance of AcceptOrderHandler.
func NewAcceptOrderHandler(p dto.AcceptOrderParams, svc services.OrderService, s bool) *AcceptOrderHandler {
	return &AcceptOrderHandler{
		params:  p,
		service: svc,
		silent:  s,
	}
}

// Handle processes accept-order command with package pricing validation, optionally suppressing output for batch import
func (h *AcceptOrderHandler) Handle() error {
	orderID := strings.TrimSpace(h.params.OrderID)
	userID := strings.TrimSpace(h.params.UserID)

	expiresAt, err := time.Parse(constants.TimeLayout, strings.TrimSpace(h.params.ExpiresAt))
	if err != nil {
		return apperrors.Newf(apperrors.ValidationFailed, "invalid expires_at format")
	}

	weight, err := handlePositiveFloatParam("weight", h.params.Weight, constants.WeightFractionDigit)
	if err != nil {
		return err
	}

	price, err := handlePositiveFloatParam("price", h.params.Price, constants.PriceFractionDigit)
	if err != nil {
		return err
	}

	pkg := normalizePackageType(h.params.Package)
	req := requests.AcceptOrderRequest{
		OrderID:   orderID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Weight:    weight,
		Price:     price,
		Package:   pkg,
	}

	order, err := h.service.AcceptOrder(req)
	if err != nil {
		return err
	}

	if h.silent {
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
