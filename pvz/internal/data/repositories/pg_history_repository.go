package repositories

import (
	"context"
	_ "embed"
	"log/slog"
	"pvz-cli/infrastructure/db"
	"pvz-cli/internal/data/queries"
	"pvz-cli/internal/data/queries/history"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

var _ HistoryRepository = (*PGHistoryRepository)(nil)

// PGHistoryRepository provides PostgreSQL-based persistence for HistoryRepository.
type PGHistoryRepository struct {
	db db.Client
}

// NewPGHistoryRepository initializes and returns a new instance of PGHistoryRepository with the provided database client.
func NewPGHistoryRepository(db db.Client) *PGHistoryRepository {
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
	countQuery, countArgs := history.BuildCountHistoryQuery(filter)
	var count int
	err := r.db.QueryRowCtx(ctx, db.ReadMode, countQuery, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}
	query, args := history.BuildFilterHistoryQuery(filter)
	rows, err := r.db.QueryCtx(ctx, db.ReadMode, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			slog.WarnContext(ctx, "rows close", "err", cerr)
		}
	}()

	var out []models.HistoryEntry
	for rows.Next() {
		var h models.HistoryEntry
		var ev string
		if err := rows.Scan(&h.OrderID, &ev, &h.Timestamp); err != nil {
			return nil, 0, err
		}
		h.Event = models.EventType(ev)
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, count, nil
}
