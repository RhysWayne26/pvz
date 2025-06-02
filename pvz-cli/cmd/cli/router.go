package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/cli/handlers"
	"strings"
)

const silentAcceptOrderOutput = false

type batchHandler func(args []string)

// Router handles command routing and execution for both batch and interactive modes
type Router struct {
	FacadeHandler handlers.FacadeHandler
	handlers      map[string]batchHandler
}

// NewRouter creates a new command router with all necessary services and command handlers
func NewRouter(
	facadeHandler handlers.FacadeHandler,
) *Router {
	r := &Router{
		FacadeHandler: facadeHandler,
		handlers:      make(map[string]batchHandler),
	}
	r.registerHandlers()
	return r
}

// Run accepts a context to allow graceful shutdown during interactive mode
func (c *Router) Run(ctx context.Context) {
	if len(os.Args) > 1 {
		c.runBatch(os.Args[1], os.Args[2:])
	} else {
		c.runInteractive(ctx)
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

func (c *Router) runInteractive(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)
	var lastScrollArgs []string

	fmt.Println("Interactive mode. Type 'help', 'exit' or commands.")
	for {
		if ctx.Err() != nil {
			fmt.Println("Shutting down...")
			return
		}

		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			reportError("read error", err)
			return
		}

		cmd, args, err := parseInteractiveLine(line, &lastScrollArgs)
		if err != nil {
			reportError("ERROR", err)
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

func reportError(prefix string, err error) {
	_, err2 := fmt.Fprintln(os.Stderr, prefix+":", err)
	if err2 != nil {
		fmt.Println(err2.Error())
	}
}

func (c *Router) registerHandlers() {
	c.handlers[constants.CmdHelp] = c.helpHandler()
	c.handlers[constants.CmdAcceptOrder] = c.acceptOrderHandler()
	c.handlers[constants.CmdReturnOrder] = c.returnOrderHandler()
	c.handlers[constants.CmdProcess] = c.processOrdersHandler()
	c.handlers[constants.CmdListOrders] = c.listOrdersHandler()
	c.handlers[constants.CmdListReturns] = c.listReturnsHandler()
	c.handlers[constants.CmdOrderHistory] = c.orderHistoryHandler()
	c.handlers[constants.CmdImportOrders] = c.importOrdersHandler()
	c.handlers[constants.CmdScrollOrders] = c.scrollOrdersHandler()
}

func (c *Router) helpHandler() batchHandler {
	return func(_ []string) {
		if err := c.FacadeHandler.HandleHelp(); err != nil {
			apperrors.Handle(err)
		}
	}
}

func (c *Router) acceptOrderHandler() batchHandler {
	return func(args []string) {
		p := NewArgsParser(args)
		params, err := p.AcceptOrderParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		if err := c.FacadeHandler.HandleAcceptOrder(params, silentAcceptOrderOutput); err != nil {
			apperrors.Handle(err)
		}
	}
}

func (c *Router) returnOrderHandler() batchHandler {
	return func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ReturnOrderParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		if err := c.FacadeHandler.HandleReturnOrder(params); err != nil {
			apperrors.Handle(err)
		}
	}
}

func (c *Router) processOrdersHandler() batchHandler {
	return func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ProcessOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		if err := c.FacadeHandler.HandleProcessOrders(params); err != nil {
			apperrors.Handle(err)
		}
	}
}

func (c *Router) listOrdersHandler() batchHandler {
	return func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ListOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		if err := c.FacadeHandler.HandleListOrders(params); err != nil {
			apperrors.Handle(err)
		}
	}
}

func (c *Router) listReturnsHandler() batchHandler {
	return func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ListReturnsParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		if err := c.FacadeHandler.HandleListReturns(params); err != nil {
			apperrors.Handle(err)
		}
	}
}

func (c *Router) orderHistoryHandler() batchHandler {
	return func(_ []string) {
		if err := c.FacadeHandler.HandleOrderHistory(); err != nil {
			apperrors.Handle(err)
		}
	}
}

func (c *Router) importOrdersHandler() batchHandler {
	return func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ImportOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		if err := c.FacadeHandler.HandleImportOrders(params); err != nil {
			apperrors.Handle(err)
		}
	}
}

func (c *Router) scrollOrdersHandler() batchHandler {
	return func(args []string) {
		p := NewArgsParser(args)
		params, err := p.ScrollOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		if err := c.FacadeHandler.HandleScrollOrders(params); err != nil {
			apperrors.Handle(err)
		}
	}
}
