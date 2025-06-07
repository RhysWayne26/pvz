package handlers

// CommandInfo contains information about a CLI command
type CommandInfo struct {
	Name        string
	Description string
	Usage       string
}

// AllCommands contains metadata for all available CLI commands
var AllCommands = []CommandInfo{
	{
		Name:        "help",
		Description: "Показать список доступных команд.",
		Usage:       "help",
	},
	{
		Name:        "accept-order",
		Description: "Принять заказ от курьера.",
		Usage:       "accept-order --order-id <id> --user-id <id> --expires <yyyy-mm-dd> --weight <float> --price <float> [--package <bag|box|film|bag+film|box+film>]",
	},
	{
		Name:        "return-order",
		Description: "Вернуть заказ курьеру.",
		Usage:       "return-order --order-id <id>",
	},
	{
		Name:        "process-orders",
		Description: "Выдать заказы или принять возврат клиента.",
		Usage:       "process-orders --user-id <id> --action <issue|return> --order-ids <id1,id2,...>",
	},
	{
		Name:        "list-orders",
		Description: "Получить список заказов.",
		Usage:       "list-orders --user-id <id> [--in-pvz] [--last-id <id>] [--last <N>] [--page <N> --limit <M>]",
	},
	{
		Name:        "list-returns",
		Description: "Получить список возвратов.",
		Usage:       "list-returns [--page <N> --limit <M>]",
	},
	{
		Name:        "order-history",
		Description: "Получить историю заказов.",
		Usage:       "order-history",
	},
	{
		Name:        "import-orders",
		Description: "Импорт заказов из JSON-файла.",
		Usage:       "import-orders --file <path>",
	},
	{
		Name:        "scroll-orders",
		Description: "Получить список заказов по принципу бесконечной прокрутки.",
		Usage:       "scroll-orders --user-id <id> [--limit <N>]",
	},
}
