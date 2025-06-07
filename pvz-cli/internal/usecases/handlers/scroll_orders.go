package handlers

import (
	"bufio"
	"fmt"
	"os"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"strings"
)

// HandleScrollOrders provides interactive CLI pagination for user orders.
func (f *DefaultFacadeHandler) HandleScrollOrders(req requests.ScrollOrdersRequest) error {
	scrollOrders(req, f.orderService)
	return nil
}

func scrollOrders(req requests.ScrollOrdersRequest, svc services.OrderService) {
	scanner := bufio.NewScanner(os.Stdin)
	var lastID uint64
	if req.LastID != nil {
		lastID = *req.LastID
	}

	for {
		orders, nextID, done, err := fetchOrdersPage(svc, req.UserID, lastID, req.Limit)
		if err != nil {
			apperrors.Handle(err)
			return
		}

		displayOrders(orders)

		if done {
			fmt.Println("NEXT: -")
			handleNoMore(scanner)
			return
		}

		fmt.Printf("NEXT: %d\n", nextID)
		lastID = nextID
		if !promptNext(scanner) {
			return
		}
	}
}

func fetchOrdersPage(
	svc services.OrderService,
	userID uint64,
	lastID uint64,
	limit *int,
) ([]models.Order, uint64, bool, error) {
	var lastIDPtr *uint64
	if lastID != 0 {
		lastIDPtr = &lastID
	}
	filter := requests.ListOrdersRequest{
		UserID: userID,
		LastID: lastIDPtr,
		Limit:  limit,
	}

	orders, nextID, _, err := svc.ListOrders(filter)
	if err != nil {
		return nil, 0, false, err
	}
	if nextID == 0 {
		return orders, 0, true, nil
	}

	return orders, nextID, false, nil
}

func displayOrders(orders []models.Order) {
	for _, o := range orders {
		fmt.Printf("ORDER: %d %d %s %s %s %.*f %.*f\n",
			o.OrderID,
			o.UserID,
			o.Status,
			o.ExpiresAt.Format(constants.TimeLayout),
			o.Package,
			constants.WeightFractionDigit, o.Weight,
			constants.PriceFractionDigit, o.Price,
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
