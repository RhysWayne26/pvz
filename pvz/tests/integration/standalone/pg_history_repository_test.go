//go:build integration

package standalone

import (
	"context"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/require"
	"pvz-cli/infrastructure/db"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/tests"
)

// TestPGHistoryRepository_Save tests saving history entries to the PostgreSQL history repository.
func TestPGHistoryRepository_Save(t *testing.T) {
	t.Parallel()
	r := runner.NewRunner(t, "PGHistoryRepository: Save")
	const orderID uint64 = 10001
	r.NewTest("Save history entries", func(t provider.T) {
		deps := newHistoryDeps(t)
		t.WithNewStep("Setup: create order", func(ctx provider.StepCtx) {
			deps.createOrder(t, orderID)
		})

		entries := []models.HistoryEntry{
			{
				OrderID:   orderID,
				Event:     models.EventAccepted,
				Timestamp: time.Now().UTC().Truncate(time.Microsecond),
			},
			{
				OrderID:   orderID,
				Event:     models.EventIssued,
				Timestamp: time.Now().UTC().Add(1 * time.Hour).Truncate(time.Microsecond),
			},
		}
		t.WithNewStep("Save accepted event", func(ctx provider.StepCtx) {
			err := deps.repo.Save(deps.ctx, entries[0])
			require.NoError(t, err)
		})
		t.WithNewStep("Save issued event", func(ctx provider.StepCtx) {
			err := deps.repo.Save(deps.ctx, entries[1])
			require.NoError(t, err)
		})

		t.WithNewStep("Verify events saved", func(ctx provider.StepCtx) {
			filter := requests.OrderHistoryFilter{
				OrderID: utils.Ptr(orderID),
				Page:    1,
				Limit:   10,
			}
			result, count, err := deps.repo.List(deps.ctx, filter)
			require.NoError(t, err)
			require.Equal(t, 2, count)
			require.Len(t, result, 2)
		})
	})

	r.RunTests()
}

// TestPGHistoryRepository_List tests the List method of PGHistoryRepository for filtering and pagination functionalities.
func TestPGHistoryRepository_List(t *testing.T) {
	t.Parallel()
	r := runner.NewRunner(t, "PGHistoryRepository: List")
	r.NewTest("List with filters", func(t provider.T) {
		deps := newHistoryDeps(t)
		const (
			orderID1 uint64 = 20001
			orderID2 uint64 = 20002
		)
		cases := []struct {
			orderID uint64
			event   models.EventType
			offset  time.Duration
		}{
			{orderID: orderID1, event: models.EventAccepted, offset: -3 * time.Hour},
			{orderID: orderID1, event: models.EventIssued, offset: -2 * time.Hour},
			{orderID: orderID2, event: models.EventAccepted, offset: -1 * time.Hour},
			{orderID: orderID2, event: models.EventReturnedByClient, offset: 0},
		}

		t.WithNewStep("Setup: create history entries", func(ctx provider.StepCtx) {
			baseTime := time.Now().UTC()
			deps.createOrder(t, orderID1)
			deps.createOrder(t, orderID2)
			for _, tc := range cases {
				entry := models.HistoryEntry{
					OrderID:   tc.orderID,
					Event:     tc.event,
					Timestamp: baseTime.Add(tc.offset).Truncate(time.Microsecond),
				}
				err := deps.repo.Save(deps.ctx, entry)
				require.NoError(t, err)
			}
		})

		t.WithNewStep("List by order ID", func(ctx provider.StepCtx) {
			filter := requests.OrderHistoryFilter{
				OrderID: utils.Ptr(orderID1),
				Page:    1,
				Limit:   10,
			}
			result, count, err := deps.repo.List(deps.ctx, filter)
			require.NoError(t, err)
			require.Equal(t, 2, count)
			require.Len(t, result, 2)
			for _, entry := range result {
				require.Equal(t, orderID1, entry.OrderID)
			}
		})

		t.WithNewStep("List all with pagination", func(ctx provider.StepCtx) {
			filter := requests.OrderHistoryFilter{
				Page:  1,
				Limit: 2,
			}
			result, count, err := deps.repo.List(deps.ctx, filter)
			require.NoError(t, err)
			require.GreaterOrEqual(t, count, 4)
			require.Len(t, result, 2)
			if len(result) >= 2 {
				require.True(t, result[0].Timestamp.After(result[1].Timestamp) ||
					result[0].Timestamp.Equal(result[1].Timestamp))
			}
		})

		t.WithNewStep("List second page", func(ctx provider.StepCtx) {
			filter := requests.OrderHistoryFilter{
				Page:  2,
				Limit: 2,
			}
			result, _, err := deps.repo.List(deps.ctx, filter)
			require.NoError(t, err)
			require.LessOrEqual(t, len(result), 2)
		})
	})

	r.RunTests()
}

func newHistoryDeps(t provider.T) historyDeps {
	commonDeps := tests.NewCommonDeps(t)
	ctx := commonDeps.Ctx
	client := commonDeps.Client
	repo := repositories.NewPGHistoryRepository(client)
	_, _ = client.ExecCtx(ctx, db.WriteMode, tests.TruncateHistorySQL)
	_, _ = client.ExecCtx(ctx, db.WriteMode, tests.TruncateOrderSql)
	return historyDeps{
		ctx:    ctx,
		client: client,
		repo:   repo,
	}
}

func (deps historyDeps) createOrder(t provider.T, orderID uint64) {
	t.Helper()
	_, err := deps.client.ExecCtx(deps.ctx, db.WriteMode,
		`INSERT INTO orders(id, user_id, status, created_at, expires_at, updated_status_at, package, weight, price, is_deleted) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false)`,
		orderID, 1, models.Accepted, time.Now(), time.Now().Add(48*time.Hour), time.Now(), models.PackageBox, 2.5, 100.0)
	require.NoError(t, err)
}

type historyDeps struct {
	ctx    context.Context
	client db.PGXClient
	repo   *repositories.PGHistoryRepository
}
