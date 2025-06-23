package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"pvz-cli/infrastructure/db"
	"pvz-cli/internal/data/queries"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

var (
	_ OrderRepository = (*PGOrderRepository)(nil)

	// ErrOrderNotFound represents an error indicating that the requested order could not be found in the database.
	ErrOrderNotFound = errors.New("order not found")
)

// PGOrderRepository provides PostgreSQL-based persistence for OrderRepository.
type PGOrderRepository struct {
	db db.Client
}

// NewPGOrderRepository initializes and returns a new instance of PGOrderRepository with the provided database client.
func NewPGOrderRepository(db db.Client) *PGOrderRepository {
	return &PGOrderRepository{
		db: db,
	}
}

// Save persists the provided order in the database.
func (r *PGOrderRepository) Save(ctx context.Context, order models.Order) error {
	_, err := r.db.ExecCtx(
		ctx,
		db.WriteMode,
		queries.SaveOrderSQL,
		order.OrderID,
		order.UserID,
		order.Status,
		order.CreatedAt,
		order.ExpiresAt,
		order.UpdatedStatusAt,
		order.Package,
		order.Weight,
		order.Price,
	)
	return err
}

// Load retrieves an order from the database by the given ID.
func (r *PGOrderRepository) Load(ctx context.Context, id uint64) (models.Order, error) {
	row := r.db.QueryRowCtx(
		ctx,
		db.ReadMode,
		queries.LoadOrderSQL,
		id,
	)
	var o models.Order
	var status, packageType string
	err := row.Scan(
		&o.OrderID,
		&o.UserID,
		&status,
		&o.CreatedAt,
		&o.ExpiresAt,
		&o.UpdatedStatusAt,
		&packageType,
		&o.Weight,
		&o.Price,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, ErrOrderNotFound
		}
		return models.Order{}, err
	}
	o.Status = models.OrderStatus(status)
	o.Package = models.PackageType(packageType)
	return o, nil
}

// Delete removes an order from the database identified by its ID.
func (r *PGOrderRepository) Delete(ctx context.Context, id uint64) error {
	res, err := r.db.ExecCtx(
		ctx, db.WriteMode,
		queries.SoftDeleteOrderSQL,
		id,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrOrderNotFound
	}
	return nil
}

// List retrieves a list of orders and the total count based on the provided filter criteria.
func (r *PGOrderRepository) List(
	ctx context.Context,
	filter requests.OrdersFilterRequest,
) ([]models.Order, int, error) {
	sqlStr, args := queries.BuildFilterOrdersQuery(filter)
	rows, err := r.db.QueryCtx(
		ctx, db.ReadMode,
		sqlStr,
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			slog.WarnContext(ctx, "rows close", "err", cerr)
		}
	}()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var st, pkg string
		if err := rows.Scan(
			&o.OrderID,
			&o.UserID,
			&st,
			&o.ExpiresAt,
			&o.Weight,
			&o.Price,
			&pkg,
		); err != nil {
			return nil, 0, err
		}
		o.Status = models.OrderStatus(st)
		o.Package = models.PackageType(pkg)
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countSQL, countArgs := queries.BuildCountOrdersQuery(filter)
	var total int
	if err := r.db.
		QueryRowCtx(
			ctx,
			db.ReadMode,
			countSQL,
			countArgs...,
		).
		Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	return orders, total, nil
}
