package cli

import (
	"bufio"
	"fmt"
	"os"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/cli/handlers"
	"pvz-cli/internal/usecases/services"
	"strings"
)

type batchHandler func(args []string)

// Router handles command routing and execution for both batch and interactive modes
type Router struct {
	OrderService   services.OrderService
	ReturnService  services.ReturnService
	HistoryService services.HistoryService

	handlers map[string]batchHandler
}

// NewRouter creates a new command router with all necessary services and command handlers
func NewRouter(
	orderSvc services.OrderService,
	returnSvc services.ReturnService,
	histSvc services.HistoryService,
) *Router {
	r := &Router{
		OrderService:   orderSvc,
		ReturnService:  returnSvc,
		HistoryService: histSvc,
		handlers:       make(map[string]batchHandler),
	}

	r.handlers[constants.CmdHelp] = func(_ []string) {
		handlers.HandleHelpCommand()
	}
	r.handlers[constants.CmdAcceptOrder] = func(args []string) {
		p := NewArgsParser(args)
		params, err := p.AcceptOrderParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleAcceptOrderCommand(params, r.OrderService)
	}
	r.handlers[constants.CmdReturnOrder] = func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ReturnOrderParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleReturnOrderCommand(params, r.ReturnService)
	}
	r.handlers[constants.CmdProcess] = func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ProcessOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleProcessOrders(params, r.OrderService, r.ReturnService)
	}
	r.handlers[constants.CmdListOrders] = func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ListOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleListOrdersCommand(params, r.OrderService)
	}
	r.handlers[constants.CmdListReturns] = func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ListReturnsParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleListReturnsCommand(params, r.ReturnService)
	}
	r.handlers[constants.CmdOrderHistory] = func(_ []string) {
		handlers.HandleOrderHistoryCommand(r.HistoryService)
	}
	r.handlers[constants.CmdImportOrders] = func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ImportOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleImportOrdersCommand(params, r.OrderService)
	}
	r.handlers[constants.CmdScrollOrders] = func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ScrollOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		handlers.HandleScrollOrdersCommand(params, r.OrderService)
	}

	return r
}

// Run starts the application in either batch mode (with command-line args) or interactive mode
func (c *Router) Run() {
	if len(os.Args) > 1 {
		c.runBatch(os.Args[1], os.Args[2:])
	} else {
		c.runInteractive()
	}
}

func (c *Router) runBatch(cmd string, args []string) {
	if h, ok := c.handlers[cmd]; ok {
		h(args)
	} else {
		fmt.Printf("ERROR: unknown command %q\n", cmd)
	}
}

func parseInteractiveLine(
	line string,
	lastScrollArgs *[]string,
) (cmd string, args []string, err error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", nil, nil
	}
	switch line {
	case constants.CmdHelp:
		return constants.CmdHelp, nil, nil
	case constants.CmdExit:
		return constants.CmdExit, nil, nil
	}
	parts := strings.Fields(line)
	cmd, args = parts[0], parts[1:]
	if cmd == constants.CmdNext {
		if *lastScrollArgs == nil {
			return "", nil, fmt.Errorf("no previous scroll-orders")
		}
		cmd, args = constants.CmdScrollOrders, *lastScrollArgs
	}
	if cmd == constants.CmdScrollOrders {
		*lastScrollArgs = args
	}
	return cmd, args, nil
}

func (c *Router) runInteractive() {
	reader := bufio.NewReader(os.Stdin)
	var lastScrollArgs []string

	fmt.Println("Interactive mode. Type 'help', 'exit' or commands.")
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			_, err2 := fmt.Fprintln(os.Stderr, "read error:", err)
			if err2 != nil {
				fmt.Println(err2.Error())
			}
			return
		}

		cmd, args, err := parseInteractiveLine(line, &lastScrollArgs)
		if err != nil {
			_, err2 := fmt.Fprintln(os.Stderr, "ERROR:", err)
			if err2 != nil {
				fmt.Println(err2.Error())
			}
			continue
		}
		if cmd == "" {
			continue
		}
		if cmd == constants.CmdExit {
			fmt.Println("Exiting...")
			return
		}
		c.runBatch(cmd, args)
	}
}
