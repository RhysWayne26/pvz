package queries

import (
	"fmt"
	"pvz-cli/internal/usecases/requests"
	"strings"
)

// SaveHistoryEntrySQL defines the SQL query for inserting a new entry into the order_history table.
const (
	SaveHistoryEntrySQL = `
insert into order_history (
	order_id,
	event,
	timestamp
) values ($1, $2, $3);
`
	historyBaseSelect = `select order_id, event, timestamp from order_history`
	historyBaseCount  = `select count(*) from order_history`
)

// BuildFilterHistoryQuery constructs a SQL query and arguments for filtering order history
func BuildFilterHistoryQuery(filter requests.OrderHistoryFilter) (string, []interface{}) {
	q, args := applyWhereForHistory(historyBaseSelect, filter)
	return applyPaginationForHistory(q, args, filter)
}

// BuildCountHistoryQuery creates a count query for history entries
func BuildCountHistoryQuery(filter requests.OrderHistoryFilter) (string, []interface{}) {
	return applyWhereForHistory(historyBaseCount, filter)
}

func applyWhereForHistory(base string, filter requests.OrderHistoryFilter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if filter.OrderID != nil {
		ph := len(args) + 1
		clauses = append(clauses, fmt.Sprintf(`order_id = $%d`, ph))
		args = append(args, *filter.OrderID)
	}
	query := base
	if len(clauses) > 0 {
		query += ` where ` + strings.Join(clauses, ` and `)
	}
	return query, args
}

func applyPaginationForHistory(query string, args []interface{}, filter requests.OrderHistoryFilter) (string, []interface{}) {
	query += ` order by timestamp desc`
	offset := (filter.Page - 1) * filter.Limit
	phLimit := len(args) + 1
	phOffset := len(args) + 2
	query += fmt.Sprintf(` limit $%d offset $%d`, phLimit, phOffset)
	args = append(args, filter.Limit, offset)
	return query, args
}
