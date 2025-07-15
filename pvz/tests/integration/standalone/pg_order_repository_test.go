//go:build integration

package standalone

import (
	"context"
	"pvz-cli/internal/infrastructure/db"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/tests"
)

// TestPGOrderRepository_SaveAndLoad validates the saving and loading functionality of the PGOrderRepository implementation.
func TestPGOrderRepository_SaveAndLoad(t *testing.T) {
	t.Parallel()
	r := runner.NewRunner(t, "PGOrderRepository: Save and Load")
	const (
		orderID uint64 = 1001
		userID  uint64 = 1
	)
	r.NewTest("Save and load order", func(t provider.T) {
		deps := newOrderDeps(t)

		t.WithNewStep("Save order", func(ctx provider.StepCtx) {
			order := models.Order{
				OrderID:         orderID,
				UserID:          userID,
				Status:          models.Accepted,
				CreatedAt:       time.Now().UTC().Truncate(time.Microsecond),
				ExpiresAt:       time.Now().UTC().Add(48 * time.Hour).Truncate(time.Microsecond),
				UpdatedStatusAt: time.Now().UTC().Truncate(time.Microsecond),
				Package:         models.PackageBox,
				Weight:          2.5,
				Price:           100.0,
			}
			err := deps.repo.Save(deps.ctx, order)
			require.NoError(t, err)
		})

		t.WithNewStep("Load and verify order", func(ctx provider.StepCtx) {
			loaded, err := deps.repo.Load(deps.ctx, orderID)
			require.NoError(t, err)
			require.Equal(t, orderID, loaded.OrderID)
			require.Equal(t, userID, loaded.UserID)
			require.Equal(t, models.Accepted, loaded.Status)
		})
	})

	r.RunTests()
}

// TestPGOrderRepository_Delete validates the delete functionality of the PGOrderRepository by checking proper deletion of an order.
func TestPGOrderRepository_Delete(t *testing.T) {
	t.Parallel()
	r := runner.NewRunner(t, "PGOrderRepository: Delete")
	const orderID uint64 = 2001

	r.NewTest("Delete order", func(t provider.T) {
		deps := newOrderDeps(t)

		t.WithNewStep("Setup: create order", func(ctx provider.StepCtx) {
			order := models.Order{
				OrderID:         orderID,
				UserID:          1,
				Status:          models.Accepted,
				CreatedAt:       time.Now().UTC().Truncate(time.Microsecond),
				ExpiresAt:       time.Now().UTC().Add(48 * time.Hour).Truncate(time.Microsecond),
				UpdatedStatusAt: time.Now().UTC().Truncate(time.Microsecond),
				Package:         models.PackageBox,
				Weight:          2.5,
				Price:           100.0,
			}
			err := deps.repo.Save(deps.ctx, order)
			require.NoError(t, err)
		})

		t.WithNewStep("Delete order", func(ctx provider.StepCtx) {
			err := deps.repo.Delete(deps.ctx, orderID)
			require.NoError(t, err)
		})

		t.WithNewStep("Verify order is deleted", func(ctx provider.StepCtx) {
			_, err := deps.repo.Load(deps.ctx, orderID)
			require.Equal(t, repositories.ErrOrderNotFound, err)
		})
	})

	r.RunTests()
}

// TestPGOrderRepository_List validates the functionality of the List method in PGOrderRepository with different filter scenarios.
func TestPGOrderRepository_List(t *testing.T) {
	t.Parallel()
	r := runner.NewRunner(t, "PGOrderRepository: List")
	r.NewTest("List with filters", func(t provider.T) {
		deps := newOrderDeps(t)
		cases := []struct {
			orderID uint64
			userID  uint64
			status  models.OrderStatus
		}{
			{orderID: 3001, userID: 10, status: models.Accepted},
			{orderID: 3002, userID: 20, status: models.Issued},
			{orderID: 3003, userID: 10, status: models.Accepted},
		}

		t.WithNewStep("Setup: create orders", func(ctx provider.StepCtx) {
			for _, tc := range cases {
				order := models.Order{
					OrderID:         tc.orderID,
					UserID:          tc.userID,
					Status:          tc.status,
					CreatedAt:       time.Now().UTC().Truncate(time.Microsecond),
					ExpiresAt:       time.Now().UTC().Add(48 * time.Hour).Truncate(time.Microsecond),
					UpdatedStatusAt: time.Now().UTC().Truncate(time.Microsecond),
					Package:         models.PackageBox,
					Weight:          2.5,
					Price:           100.0,
				}
				err := deps.repo.Save(deps.ctx, order)
				require.NoError(t, err)
			}
		})

		t.WithNewStep("List by user", func(ctx provider.StepCtx) {
			filter := requests.OrdersFilterRequest{
				UserID: utils.Ptr(uint64(10)),
			}
			result, count, err := deps.repo.List(deps.ctx, filter)
			require.NoError(t, err)
			require.GreaterOrEqual(t, count, 2)
			for _, order := range result {
				if order.OrderID >= 3001 && order.OrderID <= 3003 {
					require.Equal(t, uint64(10), order.UserID)
				}
			}
		})

		t.WithNewStep("List by status", func(ctx provider.StepCtx) {
			filter := requests.OrdersFilterRequest{
				Status: utils.Ptr(models.Accepted),
			}
			result, _, err := deps.repo.List(deps.ctx, filter)
			require.NoError(t, err)
			for _, order := range result {
				if order.OrderID >= 3001 && order.OrderID <= 3003 {
					require.Equal(t, models.Accepted, order.Status)
				}
			}
		})
	})

	r.RunTests()
}

func newOrderDeps(t provider.T) orderDeps {
	commonDeps := tests.NewCommonDeps(t)
	ctx := commonDeps.Ctx
	client := commonDeps.Client
	repo := repositories.NewPGOrderRepository(client)
	_, _ = client.ExecCtx(ctx, db.WriteMode, tests.TruncateHistorySQL)
	_, _ = client.ExecCtx(ctx, db.WriteMode, tests.TruncateOrderSql)
	return orderDeps{
		ctx:    ctx,
		client: client,
		repo:   repo,
	}
}

type orderDeps struct {
	ctx    context.Context
	client db.PGXClient
	repo   *repositories.PGOrderRepository
}
