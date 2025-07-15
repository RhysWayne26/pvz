//go:build integration

package suite

import (
	"context"
	"pvz-cli/internal/infrastructure/db"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/tests"
)

// PGOrderRepositorySuite is a testing suite for verifying the functionality of the PGOrderRepository implementation.
type PGOrderRepositorySuite struct {
	suite.Suite
}

// TestPGOrderRepositorySuite runs the test suite for the PGOrderRepository to ensure its methods function correctly.
func TestPGOrderRepositorySuite(t *testing.T) {
	t.Parallel()
	suite.RunSuite(t, new(PGOrderRepositorySuite))
}

// TestSaveAndLoad validates the saving and loading functionality of the PGOrderRepository implementation.
func (s *PGOrderRepositorySuite) TestSaveAndLoad(t provider.T) {
	const (
		orderID uint64 = 1001
		userID  uint64 = 1
	)
	deps := s.newOrderDeps(t)
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
}

// TestDelete validates the delete functionality of the PGOrderRepository by checking proper deletion of an order.
func (s *PGOrderRepositorySuite) TestDelete(t provider.T) {
	const orderID uint64 = 2001
	deps := s.newOrderDeps(t)
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
}

// TestList validates the functionality of the List method in PGOrderRepository with different filter scenarios.
func (s *PGOrderRepositorySuite) TestList(t provider.T) {
	deps := s.newOrderDeps(t)
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
}

// TestUpdate validates the update functionality of the PGOrderRepository.
func (s *PGOrderRepositorySuite) TestUpdate(t provider.T) {
	const orderID uint64 = 4001
	deps := s.newOrderDeps(t)
	t.WithNewStep("Setup: create order", func(ctx provider.StepCtx) {
		order := models.Order{
			OrderID:         orderID,
			UserID:          1,
			Status:          models.Accepted,
			CreatedAt:       time.Now().UTC().Truncate(time.Microsecond),
			ExpiresAt:       time.Now().UTC().Add(48 * time.Hour).Truncate(time.Microsecond),
			UpdatedStatusAt: time.Now().UTC().Truncate(time.Microsecond),
			Package:         models.PackageBox,
			Weight:          float32(2.5),
			Price:           float32(100.0),
		}
		err := deps.repo.Save(deps.ctx, order)
		require.NoError(t, err)
	})

	t.WithNewStep("Update order status", func(ctx provider.StepCtx) {
		updatedOrder := models.Order{
			OrderID:         orderID,
			UserID:          1,
			Status:          models.Issued,
			CreatedAt:       time.Now().UTC().Truncate(time.Microsecond),
			ExpiresAt:       time.Now().UTC().Add(48 * time.Hour).Truncate(time.Microsecond),
			UpdatedStatusAt: time.Now().UTC().Truncate(time.Microsecond),
			Package:         models.PackageBox,
			Weight:          float32(3.0),
			Price:           float32(150.0),
		}
		err := deps.repo.Save(deps.ctx, updatedOrder)
		require.NoError(t, err)
	})

	t.WithNewStep("Verify order is updated", func(ctx provider.StepCtx) {
		loaded, err := deps.repo.Load(deps.ctx, orderID)
		require.NoError(t, err)
		require.Equal(t, orderID, loaded.OrderID)
		require.Equal(t, models.Issued, loaded.Status)
		require.Equal(t, float32(3.0), loaded.Weight)
		require.Equal(t, float32(150.0), loaded.Price)
	})
}

// TestListPagination validates the pagination functionality of the List method.
func (s *PGOrderRepositorySuite) TestListPagination(t provider.T) {
	deps := s.newOrderDeps(t)

	t.WithNewStep("Setup: create multiple orders", func(ctx provider.StepCtx) {
		for i := 0; i < 5; i++ {
			order := models.Order{
				OrderID:         uint64(5001 + i),
				UserID:          1,
				Status:          models.Accepted,
				CreatedAt:       time.Now().UTC().Add(time.Duration(i) * time.Minute).Truncate(time.Microsecond),
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

	t.WithNewStep("Test first page", func(ctx provider.StepCtx) {
		filter := requests.OrdersFilterRequest{
			UserID: utils.Ptr(uint64(1)),
			Page:   utils.Ptr(1),
			Limit:  utils.Ptr(3),
		}
		result, count, err := deps.repo.List(deps.ctx, filter)
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 5)
		require.LessOrEqual(t, len(result), 3)
	})

	t.WithNewStep("Test second page", func(ctx provider.StepCtx) {
		filter := requests.OrdersFilterRequest{
			UserID: utils.Ptr(uint64(1)),
			Page:   utils.Ptr(2),
			Limit:  utils.Ptr(3),
		}
		result, count, err := deps.repo.List(deps.ctx, filter)
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 5)
		require.LessOrEqual(t, len(result), 3)
	})
}

// TestListByStatus validates filtering orders by status.
func (s *PGOrderRepositorySuite) TestListByStatus(t provider.T) {
	deps := s.newOrderDeps(t)
	t.WithNewStep("Setup: create orders with different statuses", func(ctx provider.StepCtx) {
		statuses := []models.OrderStatus{models.Accepted, models.Issued, models.Accepted, models.Issued}
		for i, status := range statuses {
			order := models.Order{
				OrderID:         uint64(6001 + i),
				UserID:          1,
				Status:          status,
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
	t.WithNewStep("List orders by accepted status", func(ctx provider.StepCtx) {
		filter := requests.OrdersFilterRequest{
			Status: utils.Ptr(models.Accepted),
		}
		result, count, err := deps.repo.List(deps.ctx, filter)
		require.NoError(t, err)
		require.GreaterOrEqual(t, count, 2)
		for _, order := range result {
			if order.OrderID >= 6001 && order.OrderID <= 6004 {
				require.Equal(t, models.Accepted, order.Status)
			}
		}
	})
}

func (s *PGOrderRepositorySuite) newOrderDeps(t provider.T) orderDeps {
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
