package constants

import "time"

// Application constants for default values, time layouts and command names
const (
	DefaultPage         = 1
	DefaultLimit        = 20
	DefaultHistoryPage  = 1
	DefaultHistoryLimit = 1000
	DefaultScrollLimit  = 20
	ReturnWindow        = 48 * time.Hour
	TimeLayout          = "2006-01-02"
	HistoryTimeLayout   = "2006-01-02 15:04:05"
	ActionIssue         = "issue"
	ActionReturn        = "return"

	CmdHelp         = "help"
	CmdAcceptOrder  = "accept-order"
	CmdReturnOrder  = "return-order"
	CmdProcess      = "process-orders"
	CmdListOrders   = "list-orders"
	CmdListReturns  = "list-returns"
	CmdOrderHistory = "order-history"
	CmdImportOrders = "import-orders"
	CmdScrollOrders = "scroll-orders"
	CmdNext         = "next"
	CmdExit         = "exit"
)
