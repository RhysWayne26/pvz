package handlers

import (
	"bufio"
	"fmt"
	"os"
	"pvz-cli/internal/utils"
	"strings"

	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
)

type ScrollOrdersParams struct {
	UserID string `json:"user_id"`
	Limit  *int   `json:"limit,omitempty"`
}

func HandleScrollOrdersCommand(params ScrollOrdersParams, svc services.OrderService) {
	userID := strings.TrimSpace(params.UserID)

	if err := utils.ValidatePositiveInt("limit", params.Limit); err != nil {
		apperrors.Handle(err)
		return
	}

	limit := constants.DefaultScrollLimit
	if params.Limit != nil {
		limit = *params.Limit
	}

	reader := bufio.NewScanner(os.Stdin)
	var lastID string
	noMoreData := false

	for {
		if noMoreData {
			fmt.Println("No more orders. Type 'exit' to quit.")
			fmt.Print("> ")
			if !reader.Scan() {
				break
			}
			cmd := strings.TrimSpace(reader.Text())
			if cmd == "exit" {
				return
			}
			fmt.Println("No more data. Only 'exit' is valid.")
			continue
		}

		filter := requests.ListOrdersFilter{
			UserID: userID,
			LastID: lastID,
			Limit:  &limit,
		}

		orders, nextID, _, err := svc.ListOrders(filter)
		if err != nil {
			apperrors.Handle(err)
			return
		}

		for _, o := range orders {
			fmt.Printf("ORDER: %s %s %s %s\n", o.OrderID, o.UserID, o.Status, o.ExpiresAt.Format(constants.TimeLayout))
		}

		if nextID != "" {
			fmt.Printf("NEXT: %s\n", nextID)
			lastID = nextID
		} else {
			fmt.Println("NEXT: -")
			noMoreData = true
		}

		fmt.Print("> ")
		if !reader.Scan() {
			break
		}
		cmd := strings.TrimSpace(reader.Text())

		switch cmd {
		case "exit":
			return
		case "next":
			continue
		default:
			fmt.Println("Type 'next' to continue or 'exit' to quit.")
		}
	}
}
