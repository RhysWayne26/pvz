package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
)

const silentAcceptOrderOutput = true

// ImportOrdersHandler processes bulk order import from a JSON file.
type ImportOrdersHandler struct {
	params  dto.ImportOrdersParams
	service services.OrderService
}

// NewImportOrdersHandler creates an instance of ImportOrdersHandler.
func NewImportOrdersHandler(p dto.ImportOrdersParams, svc services.OrderService) *ImportOrdersHandler {
	return &ImportOrdersHandler{
		params:  p,
		service: svc,
	}
}

// Handle executes the import-orders command, iterating through parsed orders
// and invoking the AcceptOrderHandler for each. Errors are printed per order.
func (h *ImportOrdersHandler) Handle() error {
	if h.params.File == "" {
		return apperrors.Newf(apperrors.ValidationFailed, "file path must not be empty")
	}

	orders, err := utils.ParseOrdersFromFile(h.params.File)
	if err != nil {
		return err
	}

	errorCount := 0
	for _, order := range orders {
		acceptHandler := NewAcceptOrderHandler(order, h.service, silentAcceptOrderOutput)
		if err := acceptHandler.Handle(); err != nil {
			fmt.Printf("ERROR importing %s: %v\n", order.OrderID, err)
			errorCount++
		}
	}

	fmt.Printf("IMPORTED: %d\n", len(orders)-errorCount)
	return nil
}
