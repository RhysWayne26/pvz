package cli

import (
	"bufio"
	"fmt"
	"os"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/cli/handlers"
	"pvz-cli/internal/usecases/services"
	"strings"
)

type Router struct {
	OrderService   services.OrderService
	ReturnService  services.ReturnService
	HistoryService services.HistoryService
}

func (c *Router) Run() {
	if len(os.Args) > 1 {
		c.runBatch(os.Args[1], os.Args[2:])
	} else {
		c.runInteractive()
	}
}

func (c *Router) runBatch(cmd string, args []string) {
	parser := NewArgsParser(args)

	switch cmd {
	case "help":
		handlers.HandleHelpCommand()

	case "accept-order":
		params, err := parser.AcceptOrderParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleAcceptOrderCommand(params, c.OrderService)

	case "return-order":
		params, err := parser.ReturnOrderParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleReturnOrderCommand(params, c.ReturnService)

	case "process-orders":
		params, err := parser.ProcessOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleProcessOrders(params, c.OrderService, c.ReturnService)

	case "list-orders":
		params, err := parser.ListOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleListOrdersCommand(params, c.OrderService)

	case "list-returns":
		params, err := parser.ListReturnsParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleListReturnsCommand(params, c.ReturnService)

	case "order-history":
		handlers.HandleOrderHistoryCommand(c.HistoryService)

	case "import-orders":
		params, err := parser.ImportOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleImportOrdersCommand(params, c.OrderService)

	case "scroll-orders":
		params, err := parser.ScrollOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleScrollOrdersCommand(params, c.OrderService)

	default:
		fmt.Printf("ERROR: unknown command %q\n", cmd)
	}
}

func (c *Router) runInteractive() {
	reader := bufio.NewReader(os.Stdin)
	var lastScrollArgs []string

	fmt.Println("Interactive mode. Type 'help', 'exit' or commands.")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			_, err := fmt.Fprintln(os.Stderr, "read error:", err)
			if err != nil {
				return
			}
			return
		}

		line = strings.TrimSpace(line)
		switch line {
		case "":
			continue
		case "help":
			handlers.HandleHelpCommand()
			continue
		case "exit":
			fmt.Println("Exiting...")
			return
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]
		if cmd == "next" {
			if lastScrollArgs == nil {
				_, err := fmt.Fprintln(os.Stderr, "ERROR: no previous scroll-orders")
				if err != nil {
					return
				}
				continue
			}
			cmd = "scroll-orders"
			args = lastScrollArgs
		}

		if cmd == "scroll-orders" {
			lastScrollArgs = args
		}

		c.runBatch(cmd, args)
	}
}
