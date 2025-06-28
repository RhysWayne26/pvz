package services_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/clock"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/builders"
	"pvz-cli/internal/usecases/mocks"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"testing"
	"time"
)

// TestDefaultOrderService_AcceptOrder_Success verifies that an order is successfully accepted with the correct flow and mocks.
func TestDefaultOrderService_AcceptOrder_Success(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)

	req := requests.AcceptOrderRequest{
		OrderID:   1,
		UserID:    42,
		Package:   models.PackageBox,
		Weight:    2.0,
		Price:     100.0,
		ExpiresAt: deps.clk.After(48 * time.Hour),
	}
	deps.repo.LoadMock.Expect(deps.ctx, req.OrderID).Return(models.Order{}, apperrors.Newf(apperrors.OrderNotFound, "order not found"))
	deps.validator.ValidateAcceptMock.Expect(models.Order{}, req).Return(nil)
	deps.pricing.EvaluateMock.Expect(req.Package, req.Weight, req.Price).Return(125.0, nil)
	deps.repo.SaveMock.Expect(deps.ctx, anyOrderFromAcceptRequest(deps.clk, req, 125.0)).Return(nil)
	deps.history.RecordMock.Expect(deps.ctx, anyEntryWith(deps.clk, req.OrderID, models.EventAccepted)).Return(nil)
	order, err := deps.svc.AcceptOrder(deps.ctx, req)
	require.NoError(t, err)
	require.Equal(t, req.OrderID, order.OrderID)
	require.Equal(t, models.Accepted, order.Status)
}

// TestAcceptOrder_FailureStages tests the AcceptOrder handler across various failure stages during the order processing.
func TestAcceptOrder_FailureStages(t *testing.T) {
	t.Parallel()
	type stage string
	const (
		stageValidate stage = "validate"
		stageEvaluate stage = "evaluate"
		stageSave     stage = "save"
		stageRecord   stage = "record"
	)
	tests := []struct {
		stage    stage
		wantCode apperrors.ErrorCode
	}{
		{stageValidate, apperrors.ValidationFailed},
		{stageEvaluate, apperrors.WeightTooHeavy},
		{stageSave, apperrors.InternalError},
		{stageRecord, apperrors.InternalError},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(string(tc.stage), func(t *testing.T) {
			t.Parallel()
			deps := newTestOrderService(t)
			req := requests.AcceptOrderRequest{
				OrderID: 1, UserID: 42,
				Package: models.PackageBox,
				Weight:  5, Price: 50,
				ExpiresAt: deps.clk.After(48 * time.Hour),
			}
			deps.repo.LoadMock.
				Expect(deps.ctx, req.OrderID).
				Return(models.Order{}, apperrors.Newf(apperrors.OrderNotFound, "not found"))
			switch tc.stage {
			case stageValidate:
				deps.validator.ValidateAcceptMock.
					Expect(models.Order{}, req).
					Return(apperrors.Newf(apperrors.ValidationFailed, "bad input"))
			default:
				deps.validator.ValidateAcceptMock.
					Expect(models.Order{}, req).
					Return(nil)
				if tc.stage == stageEvaluate {
					deps.pricing.EvaluateMock.
						Expect(req.Package, req.Weight, req.Price).
						Return(0, apperrors.Newf(apperrors.WeightTooHeavy, "too heavy"))
				} else {
					deps.pricing.EvaluateMock.
						Expect(req.Package, req.Weight, req.Price).
						Return(75.0, nil)
					if tc.stage == stageSave {
						deps.repo.SaveMock.
							Expect(deps.ctx, anyOrderFromAcceptRequest(deps.clk, req, 75.0)).
							Return(apperrors.Newf(apperrors.InternalError, "save failed"))
					} else {
						deps.repo.SaveMock.
							Expect(deps.ctx, anyOrderFromAcceptRequest(deps.clk, req, 75.0)).
							Return(nil)
						if tc.stage == stageRecord {
							deps.history.RecordMock.
								Expect(deps.ctx, anyEntryWith(deps.clk, req.OrderID, models.EventAccepted)).
								Return(apperrors.Newf(apperrors.InternalError, "record failed"))
						} else {
							deps.history.RecordMock.
								Expect(deps.ctx, anyEntryWith(deps.clk, req.OrderID, models.EventAccepted)).
								Return(nil)
						}
					}
				}
			}
			_, err := deps.svc.AcceptOrder(deps.ctx, req)
			require.Error(t, err)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			require.Equal(t, tc.wantCode, appErr.Code)
		})
	}
}

// TestDefaultOrderService_IssueOrders_Success verifies that issuing orders works as expected.
func TestDefaultOrderService_IssueOrders_Success(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)
	req := requests.IssueOrdersRequest{OrderIDs: []uint64{1, 2}}
	order1 := anyAcceptedOrder(deps.clk, 1)
	order2 := anyAcceptedOrder(deps.clk, 2)
	deps.repo.LoadMock.When(deps.ctx, uint64(1)).Then(order1, nil)
	deps.repo.LoadMock.When(deps.ctx, uint64(2)).Then(order2, nil)
	deps.validator.ValidateIssueMock.When(order1, req).Then(nil)
	deps.validator.ValidateIssueMock.When(order2, req).Then(nil)
	deps.repo.SaveMock.When(deps.ctx, anyIssuedOrder(deps.clk, 1)).Then(nil)
	deps.history.RecordMock.When(deps.ctx, anyEntryWith(deps.clk, 1, models.EventIssued)).Then(nil)
	deps.repo.SaveMock.When(deps.ctx, anyIssuedOrder(deps.clk, 2)).Then(nil)
	deps.history.RecordMock.When(deps.ctx, anyEntryWith(deps.clk, 2, models.EventIssued)).Then(nil)
	results, err := deps.svc.IssueOrders(deps.ctx, req)
	require.NoError(t, err)
	require.Len(t, results, 2)
	for _, res := range results {
		require.Nil(t, res.Error)
	}
}

// TestDefaultOrderService_IssueOrders_FailureCases ensures the IssueOrders function properly handles various failure scenarios.
func TestDefaultOrderService_IssueOrders_FailureCases(t *testing.T) {
	t.Parallel()
	type tc struct {
		name        string
		loadErr     error
		validateErr error
		saveErr     error
		wantErrCode apperrors.ErrorCode
	}
	cases := []tc{
		{
			name:        "order not found",
			loadErr:     errors.New("repo missing"),
			wantErrCode: apperrors.OrderNotFound,
		},
		{
			name:        "validation fails",
			loadErr:     nil,
			validateErr: apperrors.Newf(apperrors.ValidationFailed, "bad"),
			wantErrCode: apperrors.ValidationFailed,
		},
		{
			name:        "save fails",
			loadErr:     nil,
			validateErr: nil,
			saveErr:     apperrors.Newf(apperrors.InternalError, "db"),
			wantErrCode: apperrors.InternalError,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			deps := newTestOrderService(t)
			req := requests.IssueOrdersRequest{OrderIDs: []uint64{7}}
			var order models.Order
			deps.repo.LoadMock.
				Expect(deps.ctx, uint64(7)).
				Return(order, tc.loadErr)
			if tc.loadErr == nil {
				deps.validator.ValidateIssueMock.
					Expect(order, req).
					Return(tc.validateErr)
			}
			if tc.loadErr == nil && tc.validateErr == nil {
				upd := builders.
					NewOrderBuilderFrom(deps.clk, &order).
					WithStatus(models.Issued).
					Build()
				deps.repo.SaveMock.
					Expect(deps.ctx, upd).
					Return(tc.saveErr)
			}

			results, err := deps.svc.IssueOrders(deps.ctx, req)
			require.NoError(t, err, "service-level error should be nil")
			require.Len(t, results, 1)
			pr := results[0]
			require.Equal(t, uint64(7), pr.OrderID)
			require.Error(t, pr.Error)
			var appErr *apperrors.AppError
			require.ErrorAs(t, pr.Error, &appErr)
			require.Equal(t, tc.wantErrCode, appErr.Code)
		})
	}
}

// TestDefaultOrderService_CreateClientReturns_Success verifies that client return creation succeeds with valid inputs.
func TestDefaultOrderService_CreateClientReturns_Success(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)

	req := requests.ClientReturnsRequest{OrderIDs: []uint64{1, 2}}
	order1 := anyIssuedOrder(deps.clk, 1)
	order2 := anyIssuedOrder(deps.clk, 2)
	deps.repo.LoadMock.When(deps.ctx, uint64(1)).Then(order1, nil)
	deps.repo.LoadMock.When(deps.ctx, uint64(2)).Then(order2, nil)
	deps.validator.ValidateClientReturnMock.When(order1, req).Then(nil)
	deps.validator.ValidateClientReturnMock.When(order2, req).Then(nil)
	deps.repo.SaveMock.When(deps.ctx, anyReturnedOrder(deps.clk, 1)).Then(nil)
	deps.history.RecordMock.When(deps.ctx, anyEntryWith(deps.clk, 1, models.EventReturnedByClient)).Then(nil)
	deps.repo.SaveMock.When(deps.ctx, anyReturnedOrder(deps.clk, 2)).Then(nil)
	deps.history.RecordMock.When(deps.ctx, anyEntryWith(deps.clk, 2, models.EventReturnedByClient)).Then(nil)
	results, err := deps.svc.CreateClientReturns(deps.ctx, req)
	require.NoError(t, err)
	require.Len(t, results, 2)
	for _, res := range results {
		require.Nil(t, res.Error)
	}
}

func TestDefaultOrderService_CreateClientReturns_FailureCases(t *testing.T) {
	t.Parallel()
	type tc struct {
		name        string
		loadErr     error
		validateErr error
		saveErr     error
		wantCode    apperrors.ErrorCode
	}
	cases := []tc{
		{
			name:     "order not found",
			loadErr:  errors.New("repo missing"),
			wantCode: apperrors.OrderNotFound,
		},
		{
			name:        "validation fails",
			loadErr:     nil,
			validateErr: apperrors.Newf(apperrors.ValidationFailed, "bad"),
			wantCode:    apperrors.ValidationFailed,
		},
		{
			name:        "save fails",
			loadErr:     nil,
			validateErr: nil,
			saveErr:     apperrors.Newf(apperrors.InternalError, "db"),
			wantCode:    apperrors.InternalError,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			deps := newTestOrderService(t)
			req := requests.ClientReturnsRequest{OrderIDs: []uint64{42}}
			var order models.Order
			deps.repo.LoadMock.
				Expect(deps.ctx, uint64(42)).
				Return(order, tc.loadErr)
			if tc.loadErr == nil {
				deps.validator.ValidateClientReturnMock.
					Expect(order, req).
					Return(tc.validateErr)
			}
			if tc.loadErr == nil && tc.validateErr == nil {
				order.Status = models.Returned
				order.UpdatedStatusAt = deps.clk.Now()
				deps.repo.SaveMock.
					Expect(deps.ctx, order).
					Return(tc.saveErr)
			}

			results, err := deps.svc.CreateClientReturns(deps.ctx, req)
			require.NoError(t, err, "service-level error should be nil")
			require.Len(t, results, 1)
			pr := results[0]
			require.Equal(t, uint64(42), pr.OrderID)
			require.Error(t, pr.Error)
			var appErr *apperrors.AppError
			require.ErrorAs(t, pr.Error, &appErr)
			require.Equal(t, tc.wantCode, appErr.Code)
		})
	}
}

// TestDefaultOrderService_ReturnToCourier_Success verifies that a valid order is returned to the courier successfully.
func TestDefaultOrderService_ReturnToCourier_Success(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)

	req := requests.ReturnOrderRequest{OrderID: 99}
	order := models.Order{OrderID: 99, Status: models.Returned}
	deps.repo.LoadMock.
		Expect(deps.ctx, uint64(99)).
		Return(order, nil)
	deps.validator.ValidateReturnToCourierMock.
		Expect(order).
		Return(nil)
	deps.repo.DeleteMock.
		Expect(deps.ctx, uint64(99)).
		Return(nil)
	deps.history.RecordMock.
		Expect(deps.ctx, anyEntryWith(deps.clk, 99, models.EventReturnedToWarehouse)).
		Return(nil)

	require.NoError(t, deps.svc.ReturnToCourier(deps.ctx, req))
}

// TestDefaultOrderService_ReturnToCourier_Failures tests scenarios where the ReturnToCourier operation should fail.
func TestDefaultOrderService_ReturnToCourier_Failures(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)
	type tc struct {
		name     string
		setup    func(repo *mocks.OrderRepositoryMock, val *mocks.OrderValidatorMock, hist *mocks.HistoryServiceMock)
		wantCode apperrors.ErrorCode
	}
	cases := []tc{
		{
			name: "not found",
			setup: func(repo *mocks.OrderRepositoryMock, val *mocks.OrderValidatorMock, hist *mocks.HistoryServiceMock) {
				repo.LoadMock.Expect(deps.ctx, uint64(1)).Return(models.Order{}, errors.New("nope"))
			},
			wantCode: apperrors.OrderNotFound,
		},
		{
			name: "validation",
			setup: func(repo *mocks.OrderRepositoryMock, val *mocks.OrderValidatorMock, hist *mocks.HistoryServiceMock) {
				order := models.Order{OrderID: 1}
				repo.LoadMock.Expect(deps.ctx, uint64(1)).Return(order, nil)
				val.ValidateReturnToCourierMock.Expect(order).Return(apperrors.Newf(apperrors.ValidationFailed, "bad"))
			},
			wantCode: apperrors.ValidationFailed,
		},
		{
			name: "delete fails",
			setup: func(repo *mocks.OrderRepositoryMock, val *mocks.OrderValidatorMock, hist *mocks.HistoryServiceMock) {
				order := models.Order{OrderID: 1}
				repo.LoadMock.Expect(deps.ctx, uint64(1)).Return(order, nil)
				val.ValidateReturnToCourierMock.Expect(order).Return(nil)
				repo.DeleteMock.Expect(deps.ctx, uint64(1)).Return(errors.New("db"))
			},
			wantCode: apperrors.InternalError,
		},
		{
			name: "history record fails",
			setup: func(repo *mocks.OrderRepositoryMock, val *mocks.OrderValidatorMock, hist *mocks.HistoryServiceMock) {
				order := models.Order{OrderID: 1}
				repo.LoadMock.Expect(deps.ctx, uint64(1)).Return(order, nil)
				val.ValidateReturnToCourierMock.Expect(order).Return(nil)
				repo.DeleteMock.Expect(deps.ctx, uint64(1)).Return(nil)
				hist.RecordMock.
					Expect(deps.ctx, anyEntryWith(deps.clk, 1, models.EventReturnedToWarehouse)).
					Return(apperrors.Newf(apperrors.InternalError, "oops"))
			},
			wantCode: apperrors.InternalError,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			deps := newTestOrderService(t)
			tc.setup(deps.repo, deps.validator, deps.history)
			err := deps.svc.ReturnToCourier(deps.ctx, requests.ReturnOrderRequest{OrderID: 1})
			require.Error(t, err)
			var ae *apperrors.AppError
			require.ErrorAs(t, err, &ae)
			require.Equal(t, tc.wantCode, ae.Code)
		})
	}
}

// TestDefaultOrderService_ListOrders tests the ListOrders method of DefaultOrderService with mock dependencies and varying scenarios.
func TestDefaultOrderService_ListOrders(t *testing.T) {
	t.Parallel()
	deps := newTestOrderServiceMinimal(t)
	all := []models.Order{
		{OrderID: 1}, {OrderID: 2}, {OrderID: 3},
	}
	totalCount := 3
	deps.repo.ListMock.
		Expect(deps.ctx, requests.OrdersFilterRequest{}).
		Return(nil, 0, errors.New("fail"))
	_, _, _, err := deps.svc.ListOrders(deps.ctx, requests.OrdersFilterRequest{})
	require.Error(t, err)
	deps.repo.ListMock.
		Expect(deps.ctx, requests.OrdersFilterRequest{}).
		Return(all, totalCount, nil)
	res, next, cnt, err := deps.svc.ListOrders(deps.ctx, requests.OrdersFilterRequest{})
	require.NoError(t, err)
	require.Equal(t, all, res)
	require.Equal(t, uint64(3), next)
	require.Equal(t, 3, cnt)
	filter := requests.OrdersFilterRequest{Last: utils.Ptr(2)}
	deps.repo.ListMock.
		Expect(deps.ctx, filter).
		Return(all, totalCount, nil)
	res, next, cnt, err = deps.svc.ListOrders(deps.ctx, filter)
	require.NoError(t, err)
	require.Equal(t, all[1:], res)
	require.Equal(t, uint64(3), next)
	require.Equal(t, 2, cnt)
}

// TestDefaultOrderService_ListReturns verifies the behavior of ListReturns in DefaultOrderService.
func TestDefaultOrderService_ListReturns(t *testing.T) {
	t.Parallel()
	deps := newTestOrderServiceMinimal(t)
	expected := []models.Order{{OrderID: 5}, {OrderID: 6}}
	deps.repo.ListMock.
		Expect(deps.ctx, requests.OrdersFilterRequest{}).
		Return(nil, 0, errors.New("err"))
	_, err := deps.svc.ListReturns(deps.ctx, requests.OrdersFilterRequest{})
	require.Error(t, err)
	deps.repo.ListMock.
		Expect(deps.ctx, requests.OrdersFilterRequest{}).
		Return(expected, 0, nil)
	res, err := deps.svc.ListReturns(deps.ctx, requests.OrdersFilterRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, res)
}

// TestDefaultOrderService_CtxCancel_AcceptOrder tests the behavior of AcceptOrder when the context is canceled before execution.
func TestDefaultOrderService_CtxCancel_AcceptOrder(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := deps.svc.AcceptOrder(ctx, requests.AcceptOrderRequest{OrderID: 1})
	require.ErrorIs(t, err, context.Canceled)
}

// TestDefaultOrderService_CtxCancel_IssueOrders verifies that IssueOrders returns a context.Canceled error when context is canceled.
func TestDefaultOrderService_CtxCancel_IssueOrders(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, err := deps.svc.IssueOrders(ctx, requests.IssueOrdersRequest{OrderIDs: []uint64{1}})
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, results)
}

// TestDefaultOrderService_CtxCancel_CreateClientReturns tests the behavior of CreateClientReturns when the context is canceled.
func TestDefaultOrderService_CtxCancel_CreateClientReturns(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, err := deps.svc.CreateClientReturns(ctx, requests.ClientReturnsRequest{OrderIDs: []uint64{1}})
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, results)
}

// TestDefaultOrderService_CtxCancel_ReturnToCourier tests cancellation of context during ReturnToCourier operation.
func TestDefaultOrderService_CtxCancel_ReturnToCourier(t *testing.T) {
	t.Parallel()
	deps := newTestOrderService(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := deps.svc.ReturnToCourier(ctx, requests.ReturnOrderRequest{OrderID: 1})
	require.ErrorIs(t, err, context.Canceled)
}

// TestDefaultOrderService_CtxCancel_ListOrders verifies that ListOrders returns context.Canceled when the context is canceled.
func TestDefaultOrderService_CtxCancel_ListOrders(t *testing.T) {
	t.Parallel()
	deps := newTestOrderServiceMinimal(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, _, err := deps.svc.ListOrders(ctx, requests.OrdersFilterRequest{})
	require.ErrorIs(t, err, context.Canceled)
}

// TestDefaultOrderService_CtxCancel_ListReturns verifies ListReturns returns context.Canceled error when context is canceled.
func TestDefaultOrderService_CtxCancel_ListReturns(t *testing.T) {
	t.Parallel()
	deps := newTestOrderServiceMinimal(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := deps.svc.ListReturns(ctx, requests.OrdersFilterRequest{})
	require.ErrorIs(t, err, context.Canceled)
}

type orderSvcDeps struct {
	svc       *services.DefaultOrderService
	repo      *mocks.OrderRepositoryMock
	history   *mocks.HistoryServiceMock
	pricing   *mocks.PackagePricingServiceMock
	validator *mocks.OrderValidatorMock
	ctx       context.Context
	clk       clock.Clock
}

func newTestOrderService(t *testing.T) orderSvcDeps {
	repo := mocks.NewOrderRepositoryMock(t)
	history := mocks.NewHistoryServiceMock(t)
	pricing := mocks.NewPackagePricingServiceMock(t)
	validator := mocks.NewOrderValidatorMock(t)
	clk := &clock.FakeClock{}
	svc := services.NewDefaultOrderService(clk, repo, pricing, history, validator)
	ctx := context.Background()
	return orderSvcDeps{svc, repo, history, pricing, validator, ctx, clk}
}

type orderSvcDepsMinimal struct {
	svc  *services.DefaultOrderService
	repo *mocks.OrderRepositoryMock
	ctx  context.Context
	clk  clock.Clock
}

func newTestOrderServiceMinimal(t *testing.T) orderSvcDepsMinimal {
	repo := mocks.NewOrderRepositoryMock(t)
	clk := &clock.FakeClock{}
	svc := services.NewDefaultOrderService(clk, repo, nil, nil, nil)
	ctx := context.Background()
	return orderSvcDepsMinimal{svc: svc, repo: repo, ctx: ctx, clk: clk}
}

func anyOrderFromAcceptRequest(clk clock.Clock, req requests.AcceptOrderRequest, evaluatedPrice float32) models.Order {
	return builders.NewOrderBuilder(clk).
		WithID(req.OrderID).
		WithUserID(req.UserID).
		WithStatus(models.Accepted).
		WithPackageType(req.Package).
		WithWeight(req.Weight).
		WithPrice(evaluatedPrice).
		WithExpiresAt(req.ExpiresAt).
		Build()
}

func anyAcceptedOrder(clk clock.Clock, id uint64) models.Order {
	return builders.NewOrderBuilder(clk).
		WithID(id).
		WithStatus(models.Accepted).
		Build()
}

func anyIssuedOrder(clk clock.Clock, id uint64) models.Order {
	return builders.NewOrderBuilder(clk).
		WithID(id).
		WithStatus(models.Issued).
		Build()
}

func anyReturnedOrder(clk clock.Clock, id uint64) models.Order {
	return builders.NewOrderBuilder(clk).
		WithID(id).
		WithStatus(models.Returned).
		Build()
}

func anyEntryWith(clk clock.Clock, orderID uint64, event models.EventType) models.HistoryEntry {
	return builders.NewHistoryEntryBuilder(clk).
		WithOrderID(orderID).
		WithEvent(event).
		Build()
}
