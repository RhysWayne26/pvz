package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
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
	db db.PGXClient
}

// NewPGOrderRepository initializes and returns a new instance of PGOrderRepository with the provided database client.
func NewPGOrderRepository(db db.PGXClient) *PGOrderRepository {
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
	var order models.Order
	err := pgxscan.Get(ctx, r.db, &order, queries.LoadOrderSQL, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, ErrOrderNotFound
		}
		return models.Order{}, err
	}
	return order, nil
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
	affected := res.RowsAffected()
	if affected == 0 {
		return ErrOrderNotFound
	}
	return nil
}

// List retrieves a filtered list of orders and their total count from the database based on the specified filter criteria.
func (r *PGOrderRepository) List(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, int, error) {
	sqlStr, args := queries.BuildFilterOrdersQuery(filter)
	var orders []models.Order
	err := pgxscan.Select(ctx, r.db, &orders, sqlStr, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	countSQL, countArgs := queries.BuildCountOrdersQuery(filter)
	var total int
	err = pgxscan.Get(ctx, r.db, &total, countSQL, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}
	return orders, total, nil
}
