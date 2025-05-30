package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/services"
)

// HandleOrderHistoryCommand processes order-history command and displays all order events
func HandleOrderHistoryCommand(svc services.HistoryService) {
	entries, err := svc.ListAll(constants.DefaultHistoryPage, constants.DefaultHistoryLimit)
	if err != nil {
		apperrors.Handle(err)
		return
	}

	for _, e := range entries {
		fmt.Printf("HISTORY: %s %s %s\n", e.OrderID, e.Event, e.Timestamp.Format(constants.HistoryTimeLayout))
	}
}
