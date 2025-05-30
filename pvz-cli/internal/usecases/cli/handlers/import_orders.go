package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
)

const silentAcceptOrderOutput = true

// HandleImportOrdersCommand processes bulk order import from JSON file
func HandleImportOrdersCommand(params dto.ImportOrdersParams, svc services.OrderService) {
	if params.File == "" {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "file path must not be empty"))
		return
	}

	orders, err := utils.ParseOrdersFromFile(params.File)
	if err != nil {
		apperrors.Handle(err)
		return
	}

	errorCount := 0
	for _, order := range orders {
		if err := HandleAcceptOrderCommand(order, svc, silentAcceptOrderOutput); err != nil {
			fmt.Printf("ERROR importing %s: %v\n", order.OrderID, err)
			errorCount++
		}
	}

	fmt.Printf("IMPORTED: %d\n", len(orders)-errorCount)
}
