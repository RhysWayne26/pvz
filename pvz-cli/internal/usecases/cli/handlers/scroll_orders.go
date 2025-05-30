package handlers

import (
	"bufio"
	"fmt"
	"os"
	"pvz-cli/internal/usecases/dto"
	"strings"

	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
)

// HandleScrollOrdersCommand processes scroll-orders command with infinite scroll functionality
func HandleScrollOrdersCommand(params dto.ScrollOrdersParams, svc services.OrderService) {
	userID, limit, err := prepareScrollParams(params)
	if err != nil {
		apperrors.Handle(err)
		return
	}
	scrollOrders(userID, limit, svc)
}

func prepareScrollParams(params dto.ScrollOrdersParams) (string, int, error) {
	userID := strings.TrimSpace(params.UserID)
	if err := utils.ValidatePositiveInt("limit", params.Limit); err != nil {
		return "", 0, err
	}
	limit := constants.DefaultScrollLimit
	if params.Limit != nil {
		limit = *params.Limit
	}
	return userID, limit, nil
}

func scrollOrders(userID string, limit int, svc services.OrderService) {
	scanner := bufio.NewScanner(os.Stdin)
	lastID := ""

	for {
		orders, nextID, _, err := fetchOrdersPage(svc, userID, lastID, limit)
		if err != nil {
			apperrors.Handle(err)
			return
		}

		displayOrders(orders)
		if nextID != "" {
			fmt.Printf("NEXT: %s\n", nextID)
		} else {
			fmt.Println("NEXT: -")
			handleNoMore(scanner)
			return
		}

		lastID = nextID
		if !promptNext(scanner) {
			return
		}
	}
}

func fetchOrdersPage(
	svc services.OrderService,
	userID, lastID string,
	limit int,
) ([]models.Order, string, bool, error) {
	filter := requests.ListOrdersFilter{
		UserID: userID,
		LastID: lastID,
		Limit:  &limit,
	}
	orders, nextID, _, err := svc.ListOrders(filter)
	if err != nil {
		return nil, "", false, err
	}
	return orders, nextID, nextID == "", nil
}

func displayOrders(orders []models.Order) {
	for _, o := range orders {
		fmt.Printf("ORDER: %s %s %s %s %s %.3f %.1f\n",
			o.OrderID,
			o.UserID,
			o.Status,
			o.ExpiresAt.Format(constants.TimeLayout),
			o.Package,
			o.Weight,
			o.Price,
		)
	}

}

func promptNext(scanner *bufio.Scanner) bool {
	fmt.Print("> ")
	if !scanner.Scan() {
		return false
	}
	cmd := strings.TrimSpace(scanner.Text())
	switch cmd {
	case constants.CmdNext:
		return true
	case constants.CmdExit:
		return false
	default:
		fmt.Println("Type 'next' to continue or 'exit' to quit.")
		return promptNext(scanner)
	}
}

func handleNoMore(scanner *bufio.Scanner) {
	for {
		fmt.Println("No more orders. Type 'exit' to quit.")
		fmt.Print("> ")
		if !scanner.Scan() {
			return
		}
		if strings.TrimSpace(scanner.Text()) == constants.CmdExit {
			return
		}
		fmt.Println("No more data. Only 'exit' is valid.")
	}
}
