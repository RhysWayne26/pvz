package repositories

import (
	"context"
	_ "embed"
	"github.com/georgysavva/scany/v2/pgxscan"
	"pvz-cli/infrastructure/db"
	"pvz-cli/internal/data/queries"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

var _ HistoryRepository = (*PGHistoryRepository)(nil)

// PGHistoryRepository provides PostgreSQL-based persistence for HistoryRepository.
type PGHistoryRepository struct {
	db db.PGXClient
}

// NewPGHistoryRepository initializes and returns a new instance of PGHistoryRepository with the provided database client.
func NewPGHistoryRepository(db db.PGXClient) *PGHistoryRepository {
	return &PGHistoryRepository{
		db: db,
	}
}

// Save persists a HistoryEntry into the database.
func (r *PGHistoryRepository) Save(ctx context.Context, e models.HistoryEntry) error {
	_, err := r.db.ExecCtx(
		ctx,
		db.WriteMode,
		queries.SaveHistoryEntrySQL,
		e.OrderID,
		e.Event,
		e.Timestamp,
	)
	return err
}

// List retrieves a paginated list of all history entries from the database based on the specified page and limit.
func (r *PGHistoryRepository) List(ctx context.Context, filter requests.OrderHistoryFilter) ([]models.HistoryEntry, int, error) {
	countQuery, countArgs := queries.BuildCountHistoryQuery(filter)
	var count int
	err := pgxscan.Get(ctx, r.db, &count, countQuery, countArgs...)
	if err != nil {
		return nil, 0, err
	}
	query, args := queries.BuildFilterHistoryQuery(filter)
	var out []models.HistoryEntry
	err = pgxscan.Select(ctx, r.db, &out, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return out, count, nil
}
