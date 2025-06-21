package order

import (
	"fmt"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"strings"
)

const (

	// SelectBase defines the base SQL query for selecting columns from the orders table.
	baseSelect = "select id, user_id, status, expires_at, weight, price, package from orders"

	// CountBase defines the base SQL query for counting rows in the orders table.
	baseCount = "select count(*) from orders"
)

// BuildFilterOrdersQuery builds an SQL query string and arguments for filtering orders based on the provided filters.
func BuildFilterOrdersQuery(filter requests.OrdersFilterRequest) (string, []interface{}) {
	q, args := applyWhere(baseSelect, filter)
	return applyPagination(q, args, filter)
}

// BuildCountOrdersQuery creates a count query for orders and binds parameters based on the provided filter criteria.
func BuildCountOrdersQuery(filter requests.OrdersFilterRequest) (string, []interface{}) {
	return applyWhere(baseCount, filter)
}

func applyWhere(base string, filter requests.OrdersFilterRequest) (string, []interface{}) {
	var clauses []string
	var args []interface{}
	if filter.UserID != nil {
		ph := len(args) + 1
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", ph))
		args = append(args, *filter.UserID)
	}
	if filter.InPvz != nil && *filter.InPvz {
		ph1 := len(args) + 1
		ph2 := len(args) + 2
		clauses = append(clauses, fmt.Sprintf("status NOT IN ($%d, $%d)", ph1, ph2))
		args = append(args, models.Issued, models.Warehoused)
	}
	if filter.Status != nil {
		ph := len(args) + 1
		clauses = append(clauses, fmt.Sprintf("status = $%d", ph))
		args = append(args, *filter.Status)
	}
	if filter.LastID != nil {
		ph := len(args) + 1
		clauses = append(clauses,
			fmt.Sprintf(
				"created_at > (SELECT created_at FROM orders where id = $%d)",
				ph,
			),
		)
		args = append(args, *filter.LastID)
	}
	query := base
	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}
	return query, args
}

func applyPagination(query string, args []interface{}, filter requests.OrdersFilterRequest) (string, []interface{}) {
	// included last parameter overrides paging
	if filter.Last != nil {
		phLimit := len(args) + 1
		query = fmt.Sprintf("%s order by created_at desc limit $%d", query, phLimit)
		args = append(args, *filter.Last)
		return query, args
	}

	limit := constants.DefaultLimit
	if filter.Limit != nil {
		limit = *filter.Limit
	}
	page := constants.DefaultPage
	if filter.Page != nil {
		page = *filter.Page
	}
	offset := (page - 1) * limit
	phLimit := len(args) + 1
	phOffset := len(args) + 2
	query = fmt.Sprintf("%s order by created_at asc limit $%d offset $%d", query, phLimit, phOffset)
	args = append(args, limit, offset)
	return query, args
}
