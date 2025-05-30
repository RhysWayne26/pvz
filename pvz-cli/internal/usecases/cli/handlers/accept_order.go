package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
	"strconv"
	"strings"
	"time"
)

type AcceptOrderParams struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	ExpiresAt string `json:"expires_at"`
	Weight    string `json:"weight"`
	Price     string `json:"price"`
	Package   string `json:"package"`
}

func HandleAcceptOrderCommand(params AcceptOrderParams, svc services.OrderService) {
	orderID := strings.TrimSpace(params.OrderID)
	userID := strings.TrimSpace(params.UserID)

	expiresAt, err := time.Parse(constants.TimeLayout, strings.TrimSpace(params.ExpiresAt))
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid expires_at format"))
		return
	}

	weight, err := handlePositiveFloatParam("weight", params.Weight, 3)
	if err != nil {
		apperrors.Handle(err)
		return
	}

	price, err := handlePositiveFloatParam("price", params.Price, 1)
	if err != nil {
		apperrors.Handle(err)
		return
	}

	rawPkg := strings.TrimSpace(params.Package)
	pkg := models.PackageType(strings.TrimSpace(rawPkg))

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
		apperrors.Handle(err)
		return
	}

	fmt.Printf("ORDER_ACCEPTED: %s\n", order.OrderID)
	fmt.Printf("PACKAGE: %s\n", order.Package)
	fmt.Printf("TOTAL_PRICE: %.1f\n", order.Price)
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
