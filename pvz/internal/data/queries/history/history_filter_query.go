package history

import (
	"fmt"
	"pvz-cli/internal/usecases/requests"
	"strings"
)

const (
	baseSelect = "select order_id, event, timestamp from order_history"
	baseCount  = "select count(*) from order_history"
)

// BuildFilterHistoryQuery constructs a SQL query and arguments for filtering order history
func BuildFilterHistoryQuery(filter requests.OrderHistoryFilter) (string, []interface{}) {
	q, args := applyWhere(baseSelect, filter)
	return applyPagination(q, args, filter)
}

// BuildCountHistoryQuery creates a count query for history entries
func BuildCountHistoryQuery(filter requests.OrderHistoryFilter) (string, []interface{}) {
	return applyWhere(baseCount, filter)
}

func applyWhere(base string, filter requests.OrderHistoryFilter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if filter.OrderID != nil {
		ph := len(args) + 1
		clauses = append(clauses, fmt.Sprintf("order_id = $%d", ph))
		args = append(args, *filter.OrderID)
	}
	query := base
	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}
	return query, args
}

func applyPagination(query string, args []interface{}, filter requests.OrderHistoryFilter) (string, []interface{}) {
	query += " order by timestamp desc"
	offset := (filter.Page - 1) * filter.Limit
	phLimit := len(args) + 1
	phOffset := len(args) + 2
	query += fmt.Sprintf(" limit $%d offset $%d", phLimit, phOffset)
	args = append(args, filter.Limit, offset)
	return query, args
}
