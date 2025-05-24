package handlers

import (
	"fmt"
	"strings"

	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/services"
)

type ImportOrdersParams struct {
	File string `json:"file"`
}

func HandleImportOrdersCommand(params ImportOrdersParams, svc services.OrderService) {
	if params.File == "" {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "file path must not be empty"))
		return
	}

	count, err := svc.ImportOrders(strings.TrimSpace(params.File))
	if err != nil {
		apperrors.Handle(err)
		return
	}

	fmt.Printf("IMPORTED: %d\n", count)
}
