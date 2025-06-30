//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"pvz-cli/pkg/clock"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/require"

	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/handlers"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/usecases/services/strategies"
	"pvz-cli/internal/usecases/services/validators"
)

func TestE2E_AcceptIssueAndReturn(t *testing.T) {
	r := runner.NewRunner(t, "E2E: Accept to Issue to Return")
	const (
		orderID uint64 = 456
		userID  uint64 = 333
	)
	r.NewTest("Accept order", func(t provider.T) {
		deps := newE2EDeps(t)
		t.WithNewStep(
			fmt.Sprintf("Accept order #%d by user #%d", orderID, userID),
			func(ctx provider.StepCtx) {
				req := requests.AcceptOrderRequest{
					OrderID:   orderID,
					UserID:    userID,
					Package:   models.PackageBox,
					Weight:    1.5,
					Price:     50,
					ExpiresAt: time.Now().Add(2 * time.Hour),
				}
				res, err := deps.facade.HandleAcceptOrder(deps.ctx, req)
				require.NoError(t, err)
				require.Equal(t, orderID, res.OrderID)
				saved, err := deps.repo.Load(deps.ctx, orderID)
				require.NoError(t, err)
				require.Equal(t, models.Accepted, saved.Status)
			},
		)
	})

	r.NewTest("Issue order", func(t provider.T) {
		deps := newE2EDeps(t)
		t.WithNewStep(
			fmt.Sprintf("Setup: accept order #%d", orderID),
			func(ctx provider.StepCtx) {
				_, err := deps.facade.HandleAcceptOrder(
					deps.ctx,
					requests.AcceptOrderRequest{
						OrderID:   orderID,
						UserID:    userID,
						Package:   models.PackageBox,
						Weight:    1.5,
						Price:     50,
						ExpiresAt: time.Now().Add(2 * time.Hour),
					},
				)
				require.NoError(t, err)
			},
		)

		t.WithNewStep(
			fmt.Sprintf("Issue order #%d by user #%d", orderID, userID),
			func(ctx provider.StepCtx) {
				issueReq := requests.ProcessOrdersRequest{
					UserID:   userID,
					OrderIDs: []uint64{orderID},
					Action:   requests.ActionIssue,
				}
				issueRes, err := deps.facade.HandleProcessOrders(deps.ctx, issueReq)
				require.NoError(t, err)
				require.Contains(t, issueRes.Processed, orderID)
				issued, err := deps.repo.Load(deps.ctx, orderID)
				require.NoError(t, err)
				require.Equal(t, models.Issued, issued.Status)
			},
		)
	})

	r.NewTest("Return order", func(t provider.T) {
		deps := newE2EDeps(t)
		cases := []struct {
			orderID uint64
			userID  uint64
		}{
			{orderID: 1001, userID: 7},
			{orderID: 2002, userID: 8},
		}
		for _, tc := range cases {
			tc := tc
			t.WithNewStep(
				fmt.Sprintf("Setup: accept & issue order #%d by user #%d", tc.orderID, tc.userID),
				func(ctx provider.StepCtx) {
					_, err := deps.facade.HandleAcceptOrder(
						deps.ctx,
						requests.AcceptOrderRequest{
							OrderID:   tc.orderID,
							UserID:    tc.userID,
							Package:   models.PackageBox,
							Weight:    2.0,
							Price:     75,
							ExpiresAt: time.Now().Add(2 * time.Hour),
						},
					)
					require.NoError(t, err)
					_, err = deps.facade.HandleProcessOrders(
						deps.ctx,
						requests.ProcessOrdersRequest{
							UserID:   tc.userID,
							OrderIDs: []uint64{tc.orderID},
							Action:   requests.ActionIssue,
						},
					)
					require.NoError(t, err)
				},
			)
			t.WithNewStep(
				fmt.Sprintf("Return order #%d by user #%d", tc.orderID, tc.userID),
				func(ctx provider.StepCtx) {
					returnRes, err := deps.facade.HandleProcessOrders(
						deps.ctx,
						requests.ProcessOrdersRequest{
							UserID:   tc.userID,
							OrderIDs: []uint64{tc.orderID},
							Action:   requests.ActionReturn,
						},
					)
					require.NoError(t, err)
					require.Contains(t, returnRes.Processed, tc.orderID)
					returned, err := deps.repo.Load(deps.ctx, tc.orderID)
					require.NoError(t, err)
					require.Equal(t, models.Returned, returned.Status)
				},
			)
		}
	})
	r.RunTests()
}

type e2eDeps struct {
	ctx    context.Context
	repo   repositories.OrderRepository
	facade handlers.FacadeHandler
	store  storage.Storage
}

func newE2EDeps(t provider.T) e2eDeps {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	store := storage.NewJSONStorage(filepath.Join(dir, "data.json"))
	orderRepo := repositories.NewSnapshotOrderRepository(store)
	historyRepo := repositories.NewSnapshotHistoryRepository(store)
	clk := &clock.RealClock{}
	orderValidator := validators.NewDefaultOrderValidator(clk)
	packageValidator := validators.NewDefaultPackageValidator()
	pricingStrategy := strategies.NewDefaultPricingStrategy()
	pricingSvc := services.NewDefaultPackagePricingService(packageValidator, pricingStrategy)
	historySvc := services.NewDefaultHistoryService(historyRepo)
	orderSvc := services.NewDefaultOrderService(clk, orderRepo, pricingSvc, historySvc, orderValidator)
	facade := handlers.NewDefaultFacadeHandler(orderSvc, historySvc)
	return e2eDeps{
		ctx:    ctx,
		repo:   orderRepo,
		facade: facade,
		store:  store,
	}
}
