package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"pvz-cli/internal/usecases/requests"
	"strings"

	"pvz-cli/internal/cli/mappers"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/usecases/handlers"
)

type batchHandler func(ctx context.Context, args []string)

// Router handles command routing and execution for both batch and interactive modes
type Router struct {
	facadeHandler handlers.FacadeHandler
	facadeMapper  mappers.CLIFacadeMapper
	handlers      map[string]batchHandler
}

// NewRouter creates a new command router with all necessary services and command handlers
func NewRouter(
	facadeHandler handlers.FacadeHandler,
	facadeMapper mappers.CLIFacadeMapper,
) *Router {
	r := &Router{
		facadeHandler: facadeHandler,
		facadeMapper:  facadeMapper,
		handlers:      make(map[string]batchHandler),
	}
	r.registerHandlers()
	return r
}

// Run accepts a context to allow graceful shutdown during interactive mode
func (r *Router) Run(ctx context.Context) {
	if len(os.Args) > 1 {
		r.runBatch(ctx, os.Args[1], os.Args[2:])
	} else {
		r.runInteractive(ctx)
	}
}

func (r *Router) runBatch(ctx context.Context, cmd string, args []string) {
	if ctx.Err() != nil {
		fmt.Println("Operation cancelled")
		return
	}
	if h, ok := r.handlers[cmd]; ok {
		h(ctx, args)
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

func (r *Router) runInteractive(ctx context.Context) {
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
		r.runBatch(ctx, cmd, args)
	}
}

func reportError(prefix string, err error) {
	_, err2 := fmt.Fprintln(os.Stderr, prefix+":", err)
	if err2 != nil {
		fmt.Println(err2.Error())
	}
}

func (r *Router) registerHandlers() {
	r.handlers[constants.CmdHelp] = r.helpHandler()
	r.handlers[constants.CmdAcceptOrder] = r.acceptOrderHandler()
	r.handlers[constants.CmdReturnOrder] = r.returnOrderHandler()
	r.handlers[constants.CmdProcess] = r.processOrdersHandler()
	r.handlers[constants.CmdListOrders] = r.listOrdersHandler()
	r.handlers[constants.CmdListReturns] = r.listReturnsHandler()
	r.handlers[constants.CmdOrderHistory] = r.orderHistoryHandler()
	r.handlers[constants.CmdImportOrders] = r.importOrdersHandler()
	r.handlers[constants.CmdScrollOrders] = r.scrollOrdersHandler()
}

func (r *Router) helpHandler() batchHandler {
	return func(ctx context.Context, _ []string) {
		fmt.Println("Доступные команды:")
		for _, cmd := range AllCommands {
			fmt.Printf("  %-15s %s\n      Usage: %s\n", cmd.Name, cmd.Description, cmd.Usage)
		}
	}
}

func (r *Router) acceptOrderHandler() batchHandler {
	return func(ctx context.Context, args []string) {
		params, err := NewArgsParser(args).AcceptOrderParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		req, err := r.facadeMapper.MapAcceptOrderParams(params)
		if err != nil {
			apperrors.Handle(err)
			return
		}
		resp, err := r.facadeHandler.HandleAcceptOrder(ctx, req)
		if err != nil {
			apperrors.Handle(err)
		}
		fmt.Printf(
			"ORDER_ACCEPTED: %d\nPACKAGE: %s\nTOTAL_PRICE: %.*f\n",
			resp.OrderID,
			resp.Package,
			constants.PriceFractionDigit, resp.Price,
		)
	}
}

func (r *Router) returnOrderHandler() batchHandler {
	return func(ctx context.Context, args []string) {
		params, err := NewArgsParser(args).ReturnOrderParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		req, err := r.facadeMapper.MapReturnOrderParams(params)
		if err != nil {
			apperrors.Handle(err)
			return
		}
		res, err := r.facadeHandler.HandleReturnOrder(ctx, req)
		if err != nil {
			apperrors.Handle(err)
		}
		fmt.Printf("ORDER_RETURNED: %d\n", res.OrderID)
	}
}

func (r *Router) processOrdersHandler() batchHandler {
	return func(ctx context.Context, args []string) {
		params, err := NewArgsParser(args).ProcessOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}

		req, err := r.facadeMapper.MapProcessOrdersParams(params)
		if err != nil {
			apperrors.Handle(err)
			return
		}

		resp, err := r.facadeHandler.HandleProcessOrders(ctx, req)
		if err != nil {
			apperrors.Handle(err)
			return
		}

		for _, id := range resp.Processed {
			fmt.Printf("PROCESSED: %d\n", id)
		}

		for _, report := range resp.Failed {
			apperrors.Handle(report.Error)
		}
	}
}

func (r *Router) listOrdersHandler() batchHandler {
	return func(ctx context.Context, args []string) {
		params, err := NewArgsParser(args).ListOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		req, err := r.facadeMapper.MapListOrdersParams(params)
		if err != nil {
			apperrors.Handle(err)
			return
		}
		res, err := r.facadeHandler.HandleListOrders(ctx, req)
		if err != nil {
			apperrors.Handle(err)
		}

		for _, o := range res.Orders {
			fmt.Printf(
				"ORDER: %d %d %s %s %s %.*f %.*f\n",
				o.OrderID, o.UserID, o.Status,
				o.ExpiresAt.Format(constants.TimeLayout),
				o.Package,
				constants.WeightFractionDigit, o.Weight,
				constants.PriceFractionDigit, o.Price,
			)
		}
		if res.Total != nil {
			fmt.Printf("TOTAL: %d\n", *res.Total)
		}
	}
}

func (r *Router) listReturnsHandler() batchHandler {
	return func(ctx context.Context, args []string) {
		params, err := NewArgsParser(args).ListReturnsParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}

		req, err := r.facadeMapper.MapListReturnsParams(params)
		if err != nil {
			apperrors.Handle(err)
			return
		}
		res, err := r.facadeHandler.HandleListOrders(ctx, req)
		if err != nil {
			apperrors.Handle(err)
			return
		}
		for _, o := range res.Orders {
			fmt.Printf(
				"ORDER: %d %s %s %.*f\n",
				o.OrderID, o.Status, o.Package,
				constants.PriceFractionDigit, o.Price,
			)
		}
		fmt.Printf("PAGE: %d LIMIT: %d\n", *req.Page, *req.Limit)

	}
}

func (r *Router) orderHistoryHandler() batchHandler {
	return func(ctx context.Context, args []string) {
		params, err := NewArgsParser(args).OrderHistoryParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		req, err := r.facadeMapper.MapOrderHistoryParams(params)
		if err != nil {
			apperrors.Handle(err)
			return
		}
		res, err := r.facadeHandler.HandleOrderHistory(ctx, req)
		if err != nil {
			apperrors.Handle(err)
		}
		for _, e := range res.History {
			fmt.Printf("HISTORY: %d %s %s\n",
				e.OrderID,
				e.Event,
				e.Timestamp.Format(constants.HistoryTimeLayout),
			)
		}
	}
}

func (r *Router) importOrdersHandler() batchHandler {
	return func(ctx context.Context, args []string) {
		params, err := NewArgsParser(args).ImportOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}
		req, err := r.facadeMapper.MapImportOrdersParams(params)
		if err != nil {
			apperrors.Handle(err)
		}

		res, err := r.facadeHandler.HandleImportOrders(ctx, req)
		if err != nil {
			apperrors.Handle(err)
			return
		}

		for _, status := range res.Statuses {
			if status.Error != nil {
				apperrors.Handle(status.Error)
			}
		}

		fmt.Printf("IMPORTED: %d\n", res.Imported)
	}
}

func (r *Router) scrollOrdersHandler() batchHandler {
	return func(ctx context.Context, args []string) {
		cliParams, err := NewArgsParser(args).ScrollOrdersParams()
		if err != nil {
			apperrors.Handle(err)
			return
		}

		req, err := r.facadeMapper.MapScrollOrdersParams(cliParams)
		if err != nil {
			apperrors.Handle(err)
			return
		}

		scanner := bufio.NewScanner(os.Stdin)
		r.runScrollLoop(ctx, req, scanner)
	}
}

func (r *Router) runScrollLoop(ctx context.Context, req requests.OrdersFilterRequest, scanner *bufio.Scanner) {
	for {
		resp, err := r.facadeHandler.HandleListOrders(ctx, req)
		if err != nil {
			apperrors.Handle(err)
			return
		}

		for _, o := range resp.Orders {
			fmt.Printf("ORDER: %d %d %s %s %s %.*f %.*f\n",
				o.OrderID, o.UserID, o.Status,
				o.ExpiresAt.Format(constants.TimeLayout),
				o.Package,
				constants.WeightFractionDigit, o.Weight,
				constants.PriceFractionDigit, o.Price,
			)
		}

		if resp.NextID == nil || *resp.NextID == 0 {
			fmt.Println("NEXT: -")
			r.waitForExit(ctx, scanner)
			return
		}

		fmt.Printf("NEXT: %d\n", *resp.NextID)
		req.LastID = resp.NextID

		if !promptNext(scanner) {
			return
		}
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

func (r *Router) waitForExit(ctx context.Context, scanner *bufio.Scanner) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down...")
			return
		default:
		}

		fmt.Println("No more orders. Type 'exit' to quit.")
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		if strings.TrimSpace(scanner.Text()) == constants.CmdExit {
			return
		}
	}
}
