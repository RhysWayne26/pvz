package queries

import (
	"fmt"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"strings"
)

const (
	// SaveOrderSQL is a SQL query for inserting or updating an order in the orders table, using ON CONFLICT for upserts.
	SaveOrderSQL = `
insert into orders(
                   id,
                   user_id,
                   status,
                   created_at,
                   expires_at,
                   updated_status_at,
                   package,
                   weight,
                   price)
values (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
)
on conflict (id) do update set
user_id            = EXCLUDED.user_id,
status             = EXCLUDED.status,
created_at         = LEAST(orders.created_at, EXCLUDED.created_at),
expires_at         = EXCLUDED.expires_at,
updated_status_at  = EXCLUDED.updated_status_at,
package            = EXCLUDED.package,
weight             = EXCLUDED.weight,
price              = EXCLUDED.price;
`
	// LoadOrderSQL is the SQL query to retrieve non-deleted order details by order ID from the 'orders' table.
	LoadOrderSQL = `
select id,
	user_id,
	status,
	created_at,
	expires_at,
	updated_status_at,
	package,
	weight,
	price
from orders
where id = $1 and is_deleted = false;
`

	// SoftDeleteOrderSQL is a SQL query to mark an order as deleted in the 'orders' table by setting is_deleted to true.
	SoftDeleteOrderSQL = `
update orders
	set is_deleted = true
where id = $1;
`
	orderBaseSelect = `select id, user_id, status, expires_at, weight, price, package from orders`
	orderBaseCount  = `select count(*) from orders`
)

// BuildFilterOrdersQuery builds an SQL query string and arguments for filtering orders based on the provided filters.
func BuildFilterOrdersQuery(filter requests.OrdersFilterRequest) (string, []interface{}) {
	q, args := applyWhereForOrders(orderBaseSelect, filter)
	return applyPaginationForOrders(q, args, filter)
}

// BuildCountOrdersQuery creates a count query for orders and binds parameters based on the provided filter criteria.
func BuildCountOrdersQuery(filter requests.OrdersFilterRequest) (string, []interface{}) {
	return applyWhereForOrders(orderBaseCount, filter)
}

func applyWhereForOrders(base string, filter requests.OrdersFilterRequest) (string, []interface{}) {
	var clauses []string
	clauses = append(clauses, `is_deleted = false`)
	var args []interface{}
	if filter.UserID != nil {
		ph := len(args) + 1
		clauses = append(clauses, fmt.Sprintf(`user_id = $%d`, ph))
		args = append(args, *filter.UserID)
	}
	if filter.InPvz != nil && *filter.InPvz {
		ph := len(args) + 1
		clauses = append(clauses, fmt.Sprintf(`status <> $%d`, ph))
		args = append(args, models.Issued)
	}
	if filter.Status != nil {
		ph := len(args) + 1
		clauses = append(clauses, fmt.Sprintf(`status = $%d`, ph))
		args = append(args, *filter.Status)
	}
	if filter.LastID != nil {
		ph := len(args) + 1
		clauses = append(clauses,
			fmt.Sprintf(
				`created_at > (select created_at from orders where id = $%d)`,
				ph,
			),
		)
		args = append(args, *filter.LastID)
	}
	query := base
	if len(clauses) > 0 {
		query += ` where ` + strings.Join(clauses, ` and `)
	}
	return query, args
}

func applyPaginationForOrders(query string, args []interface{}, filter requests.OrdersFilterRequest) (string, []interface{}) {
	// included last parameter overrides paging
	if filter.Last != nil {
		phLimit := len(args) + 1
		query = fmt.Sprintf(`%s order by created_at desc limit $%d`, query, phLimit)
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
	query = fmt.Sprintf(`%s order by created_at asc limit $%d offset $%d`, query, phLimit, phOffset)
	args = append(args, limit, offset)
	return query, args
}
