package cli

import (
	"bufio"
	"fmt"
	"os"
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
		handlers.HandleAcceptOrderCommand(parser.AcceptOrderParams(), c.OrderService)
	case "return-order":
		handlers.HandleReturnOrderCommand(parser.ReturnOrderParams(), c.ReturnService)
	case "process-orders":
		handlers.HandleProcessOrders(parser.ProcessOrdersParams(), c.OrderService, c.ReturnService)
	case "list-orders":
		handlers.HandleListOrdersCommand(parser.ListOrdersParams(), c.OrderService)
	case "list-returns":
		handlers.HandleListReturnsCommand(parser.ListReturnsParams(), c.ReturnService)
	case "order-history":
		handlers.HandleOrderHistoryCommand(c.HistoryService)
	case "import-orders":
		handlers.HandleImportOrdersCommand(parser.ImportOrdersParams(), c.OrderService)
	case "scroll-orders":
		handlers.HandleScrollOrdersCommand(parser.ScrollOrdersParams(), c.OrderService)
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
