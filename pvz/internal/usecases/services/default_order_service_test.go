package services

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"pvz-cli/infrastructure/db"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/utils"
	repmocks "pvz-cli/internal/data/repositories/mocks"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	svcmocks "pvz-cli/internal/usecases/services/mocks"
	valmocks "pvz-cli/internal/usecases/services/validators/mocks"
	"pvz-cli/pkg/clock"
	"pvz-cli/tests/builders"
	"testing"
	"time"
)

type SyncPoolStub struct {
	shutdown bool
}

func (p *SyncPoolStub) Submit(task func()) {
	task()
}

func (p *SyncPoolStub) SubmitWithContext(ctx context.Context, task func()) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	task()
	return nil
}

func (p *SyncPoolStub) Shutdown() {
	p.shutdown = true
}

func (p *SyncPoolStub) ShutdownWithTimeout(timeout time.Duration) error {
	p.shutdown = true
	return nil
}

func (p *SyncPoolStub) SetWorkerCount(count int) {
}

func (p *SyncPoolStub) GetStats() map[string]interface{} {
	return map[string]interface{}{}
}

func (p *SyncPoolStub) IsShutdown() bool {
	return p.shutdown
}

type acceptanceStage string

const (
	stageValidate acceptanceStage = "validate"
	stageEvaluate acceptanceStage = "evaluate"
	stageSave     acceptanceStage = "save"
	stageRecord   acceptanceStage = "record"
	stageOutbox   acceptanceStage = "outbox"
)

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
	deps.actorSvc.DetermineActorMock.Expect(deps.ctx, models.EventAccepted, req.UserID).Return(models.Actor{}, nil)
	deps.repo.SaveMock.Set(func(ctx context.Context, order models.Order) error {
		require.Equal(t, req.OrderID, order.OrderID)
		require.Equal(t, models.Accepted, order.Status)
		require.Equal(t, float32(125.0), order.Price)
		return nil
	})
	deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
		require.Greater(t, len(payload), 0)
		require.Contains(t, string(payload), "order_accepted")
		return nil
	})

	deps.history.RecordMock.Set(func(ctx context.Context, entry models.HistoryEntry) error {
		require.Equal(t, req.OrderID, entry.OrderID)
		require.Equal(t, models.EventAccepted, entry.Event)
		return nil
	})
	order, err := deps.svc.AcceptOrder(deps.ctx, req)
	require.NoError(t, err)
	require.Equal(t, req.OrderID, order.OrderID)
	require.Equal(t, models.Accepted, order.Status)
}

func TestDefaultOrderService_AcceptOrder_FailureCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name            string
		acceptanceStage acceptanceStage
		mockErr         error
		wantCode        apperrors.ErrorCode
	}{
		{"validation fails", stageValidate, apperrors.Newf(apperrors.ValidationFailed, "bad input"), apperrors.ValidationFailed},
		{"evaluation fails", stageEvaluate, apperrors.Newf(apperrors.WeightTooHeavy, "too heavy"), apperrors.WeightTooHeavy},
		{"save fails", stageSave, apperrors.Newf(apperrors.InternalError, "save fail"), apperrors.InternalError},
		{"outbox fails", stageOutbox, apperrors.Newf(apperrors.InternalError, "outbox fail"), apperrors.InternalError},
		{"record fails", stageRecord, apperrors.Newf(apperrors.InternalError, "record fail"), apperrors.InternalError},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			deps := newTestOrderService(t)
			req := newAcceptOrderRequest(
				1, models.PackageBox, 5, deps.clk.After(48*time.Hour),
			)
			deps.repo.LoadMock.
				Expect(deps.ctx, req.OrderID).
				Return(models.Order{}, apperrors.Newf(apperrors.OrderNotFound, "not found"))
			mockAcceptFailure(deps, req, tc.acceptanceStage, tc.mockErr)
			_, err := deps.svc.AcceptOrder(deps.ctx, req)
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
	order1 := builders.NewOrderBuilder(deps.clk).
		WithID(1).
		WithUserID(42).
		WithStatus(models.Accepted).
		Build()
	order2 := builders.NewOrderBuilder(deps.clk).
		WithID(2).
		WithUserID(43).
		WithStatus(models.Accepted).
		Build()
	deps.repo.LoadMock.When(deps.ctx, uint64(1)).Then(order1, nil)
	deps.repo.LoadMock.When(deps.ctx, uint64(2)).Then(order2, nil)
	deps.validator.ValidateIssueMock.When(order1, req).Then(nil)
	deps.validator.ValidateIssueMock.When(order2, req).Then(nil)
	actorCallCount := 0
	deps.actorSvc.DetermineActorMock.Set(func(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error) {
		actorCallCount++
		require.Equal(t, models.EventIssued, event)
		require.Contains(t, []uint64{42, 43}, userID)
		return models.Actor{}, nil
	})
	saveCallCount := 0
	deps.repo.SaveMock.Set(func(ctx context.Context, order models.Order) error {
		saveCallCount++
		require.Equal(t, models.Issued, order.Status)
		require.Contains(t, []uint64{1, 2}, order.OrderID)
		return nil
	})
	outboxCallCount := 0
	deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
		outboxCallCount++
		require.Greater(t, len(payload), 0)
		require.Contains(t, string(payload), "order_issued")
		return nil
	})
	historyCallCount := 0
	deps.history.RecordMock.Set(func(ctx context.Context, entry models.HistoryEntry) error {
		historyCallCount++
		require.Equal(t, models.EventIssued, entry.Event)
		require.Contains(t, []uint64{1, 2}, entry.OrderID)
		return nil
	})
	results, err := deps.svc.IssueOrders(deps.ctx, req)
	require.NoError(t, err)
	require.Len(t, results, 2)
	for _, res := range results {
		require.Nil(t, res.Error)
	}
	require.Equal(t, 2, actorCallCount)
	require.Equal(t, 2, saveCallCount)
	require.Equal(t, 2, outboxCallCount)
	require.Equal(t, 2, historyCallCount)
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
			order := models.Order{
				OrderID: 7,
				UserID:  42,
				Status:  models.Accepted,
			}
			deps.repo.LoadMock.
				Expect(deps.ctx, uint64(7)).
				Return(order, tc.loadErr)
			if tc.loadErr == nil {
				deps.validator.ValidateIssueMock.
					Expect(order, req).
					Return(tc.validateErr)
			}
			if tc.loadErr == nil && tc.validateErr == nil {
				deps.actorSvc.DetermineActorMock.Set(func(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error) {
					require.Equal(t, models.EventIssued, event)
					require.Equal(t, uint64(42), userID)
					return models.Actor{}, nil
				})
				deps.repo.SaveMock.Set(func(ctx context.Context, savedOrder models.Order) error {
					require.Equal(t, models.Issued, savedOrder.Status)
					require.Equal(t, uint64(7), savedOrder.OrderID)
					return tc.saveErr
				})
				if tc.saveErr == nil {
					deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
						return nil
					})
					deps.history.RecordMock.Set(func(ctx context.Context, entry models.HistoryEntry) error {
						return nil
					})
				}
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
	order1 := builders.NewOrderBuilder(deps.clk).
		WithID(1).
		WithUserID(42).
		WithStatus(models.Issued).
		Build()
	order2 := builders.NewOrderBuilder(deps.clk).
		WithID(2).
		WithUserID(43).
		WithStatus(models.Issued).
		Build()
	deps.repo.LoadMock.When(deps.ctx, uint64(1)).Then(order1, nil)
	deps.repo.LoadMock.When(deps.ctx, uint64(2)).Then(order2, nil)
	deps.validator.ValidateClientReturnMock.When(order1, req).Then(nil)
	deps.validator.ValidateClientReturnMock.When(order2, req).Then(nil)
	actorCallCount := 0
	deps.actorSvc.DetermineActorMock.Set(func(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error) {
		actorCallCount++
		require.Equal(t, models.EventReturnedByClient, event)
		require.Contains(t, []uint64{42, 43}, userID)
		return models.Actor{}, nil
	})
	saveCallCount := 0
	deps.repo.SaveMock.Set(func(ctx context.Context, order models.Order) error {
		saveCallCount++
		require.Equal(t, models.Returned, order.Status)
		require.Contains(t, []uint64{1, 2}, order.OrderID)
		return nil
	})
	outboxCallCount := 0
	deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
		outboxCallCount++
		require.Greater(t, len(payload), 0)
		require.Contains(t, string(payload), "order_returned_by_client")
		return nil
	})
	historyCallCount := 0
	deps.history.RecordMock.Set(func(ctx context.Context, entry models.HistoryEntry) error {
		historyCallCount++
		require.Equal(t, models.EventReturnedByClient, entry.Event)
		require.Contains(t, []uint64{1, 2}, entry.OrderID)
		return nil
	})
	results, err := deps.svc.CreateClientReturns(deps.ctx, req)
	require.NoError(t, err)
	require.Len(t, results, 2)
	for _, res := range results {
		require.Nil(t, res.Error)
	}
	require.Equal(t, 2, actorCallCount)
	require.Equal(t, 2, saveCallCount)
	require.Equal(t, 2, outboxCallCount)
	require.Equal(t, 2, historyCallCount)
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
			order := models.Order{
				OrderID: 42,
				UserID:  123,
				Status:  models.Issued,
			}
			deps.repo.LoadMock.
				Expect(deps.ctx, uint64(42)).
				Return(order, tc.loadErr)
			if tc.loadErr == nil {
				deps.validator.ValidateClientReturnMock.
					Expect(order, req).
					Return(tc.validateErr)
			}
			if tc.loadErr == nil && tc.validateErr == nil {
				deps.actorSvc.DetermineActorMock.Set(func(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error) {
					require.Equal(t, models.EventReturnedByClient, event)
					require.Equal(t, uint64(123), userID)
					return models.Actor{}, nil
				})
				deps.repo.SaveMock.Set(func(ctx context.Context, savedOrder models.Order) error {
					require.Equal(t, models.Returned, savedOrder.Status)
					require.Equal(t, uint64(42), savedOrder.OrderID)
					return tc.saveErr
				})
				if tc.saveErr == nil {
					deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
						return nil
					})
					deps.history.RecordMock.Set(func(ctx context.Context, entry models.HistoryEntry) error {
						return nil
					})
				}
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
	order := models.Order{
		OrderID: 99,
		UserID:  42,
		Status:  models.Returned,
	}
	deps.repo.LoadMock.
		Expect(deps.ctx, uint64(99)).
		Return(order, nil)
	deps.validator.ValidateReturnToCourierMock.
		Expect(order).
		Return(nil)
	deps.actorSvc.DetermineActorMock.Set(func(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error) {
		require.Equal(t, models.EventReturnedToWarehouse, event)
		require.Equal(t, uint64(42), userID)
		return models.Actor{}, nil
	})
	deleteCallCount := 0
	deps.repo.DeleteMock.Set(func(ctx context.Context, orderID uint64) error {
		deleteCallCount++
		require.Equal(t, uint64(99), orderID)
		return nil
	})
	outboxCallCount := 0
	deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
		outboxCallCount++
		require.Greater(t, len(payload), 0)
		require.Contains(t, string(payload), "order_returned_to_courier")
		return nil
	})
	historyCallCount := 0
	deps.history.RecordMock.Set(func(ctx context.Context, entry models.HistoryEntry) error {
		historyCallCount++
		require.Equal(t, models.EventReturnedToWarehouse, entry.Event)
		require.Equal(t, uint64(99), entry.OrderID)
		return nil
	})
	require.NoError(t, deps.svc.ReturnToCourier(deps.ctx, req))
	require.Equal(t, 1, deleteCallCount)
	require.Equal(t, 1, outboxCallCount)
	require.Equal(t, 1, historyCallCount)
}

// TestDefaultOrderService_ReturnToCourier_Failures tests scenarios where the ReturnToCourier operation should fail.
func TestDefaultOrderService_ReturnToCourier_Failures(t *testing.T) {
	t.Parallel()
	type tc struct {
		name     string
		setup    func(deps orderSvcDeps)
		wantCode apperrors.ErrorCode
	}
	cases := []tc{
		{
			name: "not found",
			setup: func(deps orderSvcDeps) {
				deps.repo.LoadMock.Expect(deps.ctx, uint64(1)).Return(models.Order{}, errors.New("nope"))
			},
			wantCode: apperrors.OrderNotFound,
		},
		{
			name: "validation",
			setup: func(deps orderSvcDeps) {
				order := models.Order{OrderID: 1, UserID: 42}
				deps.repo.LoadMock.Expect(deps.ctx, uint64(1)).Return(order, nil)
				deps.validator.ValidateReturnToCourierMock.Expect(order).Return(apperrors.Newf(apperrors.ValidationFailed, "bad"))
			},
			wantCode: apperrors.ValidationFailed,
		},
		{
			name: "delete fails",
			setup: func(deps orderSvcDeps) {
				order := models.Order{OrderID: 1, UserID: 42}
				deps.repo.LoadMock.Expect(deps.ctx, uint64(1)).Return(order, nil)
				deps.validator.ValidateReturnToCourierMock.Expect(order).Return(nil)
				deps.actorSvc.DetermineActorMock.Set(func(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error) {
					return models.Actor{}, nil
				})
				deps.repo.DeleteMock.Set(func(ctx context.Context, orderID uint64) error {
					return errors.New("db")
				})
			},
			wantCode: apperrors.InternalError,
		},
		{
			name: "history record fails",
			setup: func(deps orderSvcDeps) {
				order := models.Order{OrderID: 1, UserID: 42}
				deps.repo.LoadMock.Expect(deps.ctx, uint64(1)).Return(order, nil)
				deps.validator.ValidateReturnToCourierMock.Expect(order).Return(nil)
				deps.actorSvc.DetermineActorMock.Set(func(ctx context.Context, event models.EventType, userID uint64) (models.Actor, error) {
					return models.Actor{}, nil
				})
				deps.repo.DeleteMock.Set(func(ctx context.Context, orderID uint64) error {
					return nil
				})
				deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
					return nil
				})
				deps.history.RecordMock.Set(func(ctx context.Context, entry models.HistoryEntry) error {
					return apperrors.Newf(apperrors.InternalError, "oops")
				})
			},
			wantCode: apperrors.InternalError,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			deps := newTestOrderService(t)
			tc.setup(deps)
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
	svc        *DefaultOrderService
	repo       *repmocks.OrderRepositoryMock
	outboxRepo *repmocks.OutboxRepositoryMock
	history    *svcmocks.HistoryServiceMock
	pricing    *svcmocks.PackagePricingServiceMock
	actorSvc   *svcmocks.ActorServiceMock
	validator  *valmocks.OrderValidatorMock
	txRunner   *db.NoOpTxRunner
	ctx        context.Context
	clk        clock.Clock
}

func newTestOrderService(t *testing.T) orderSvcDeps {
	repo := repmocks.NewOrderRepositoryMock(t)
	outboxRepo := repmocks.NewOutboxRepositoryMock(t)
	history := svcmocks.NewHistoryServiceMock(t)
	pricing := svcmocks.NewPackagePricingServiceMock(t)
	actorSvc := svcmocks.NewActorServiceMock(t)
	validator := valmocks.NewOrderValidatorMock(t)
	txRunner := db.NewNoOpTxRunner()
	clk := &clock.FakeClock{}
	pool := &SyncPoolStub{}
	ctx := context.Background()
	svc := NewDefaultOrderService(clk, pool, txRunner, repo, outboxRepo, pricing, history, actorSvc, validator)
	return orderSvcDeps{svc, repo, outboxRepo, history, pricing, actorSvc, validator, txRunner, ctx, clk}
}

type orderSvcDepsMinimal struct {
	svc  *DefaultOrderService
	repo *repmocks.OrderRepositoryMock
	ctx  context.Context
	clk  clock.Clock
}

func newTestOrderServiceMinimal(t *testing.T) orderSvcDepsMinimal {
	repo := repmocks.NewOrderRepositoryMock(t)
	clk := &clock.FakeClock{}
	pool := &SyncPoolStub{}
	txRunner := db.NewNoOpTxRunner()
	svc := NewDefaultOrderService(clk, pool, txRunner, repo, nil, nil, nil, nil, nil)
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

// User ID and price are irrelevant for validator at this point
func newAcceptOrderRequest(id uint64, packageType models.PackageType, weight float32, expiresAt time.Time) requests.AcceptOrderRequest {
	return requests.AcceptOrderRequest{
		OrderID:   id,
		UserID:    42,
		Package:   packageType,
		Weight:    weight,
		Price:     100,
		ExpiresAt: expiresAt,
	}
}

func mockAcceptFailureBase(deps orderSvcDeps, req requests.AcceptOrderRequest) {
	deps.repo.LoadMock.
		Expect(deps.ctx, req.OrderID).
		Return(models.Order{}, apperrors.Newf(apperrors.OrderNotFound, "not found"))
}

func mockAcceptFailureValidation(deps orderSvcDeps, req requests.AcceptOrderRequest, mockErr error) {
	mockAcceptFailureBase(deps, req)
	deps.validator.ValidateAcceptMock.
		Expect(models.Order{}, req).
		Return(mockErr)
}

func mockAcceptFailureEvaluation(deps orderSvcDeps, req requests.AcceptOrderRequest, mockErr error) {
	mockAcceptFailureBase(deps, req)
	deps.validator.ValidateAcceptMock.
		Expect(models.Order{}, req).
		Return(nil)
	deps.pricing.EvaluateMock.
		Expect(req.Package, req.Weight, req.Price).
		Return(0, mockErr)
}

func mockAcceptFailureSave(deps orderSvcDeps, req requests.AcceptOrderRequest, mockErr error) {
	mockAcceptFailureBase(deps, req)
	deps.validator.ValidateAcceptMock.
		Expect(models.Order{}, req).
		Return(nil)
	deps.pricing.EvaluateMock.
		Expect(req.Package, req.Weight, req.Price).
		Return(75.0, nil)
	deps.actorSvc.DetermineActorMock.
		Expect(deps.ctx, models.EventAccepted, req.UserID).
		Return(models.Actor{}, nil)
	deps.repo.SaveMock.Set(func(ctx context.Context, order models.Order) error {
		return mockErr
	})
}

func mockAcceptFailureOutbox(deps orderSvcDeps, req requests.AcceptOrderRequest, mockErr error) {
	mockAcceptFailureBase(deps, req)
	deps.validator.ValidateAcceptMock.
		Expect(models.Order{}, req).
		Return(nil)
	deps.pricing.EvaluateMock.
		Expect(req.Package, req.Weight, req.Price).
		Return(75.0, nil)
	deps.actorSvc.DetermineActorMock.
		Expect(deps.ctx, models.EventAccepted, req.UserID).
		Return(models.Actor{}, nil)
	deps.repo.SaveMock.Set(func(ctx context.Context, order models.Order) error {
		return nil
	})
	deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
		return mockErr
	})
}

func mockAcceptFailureRecord(deps orderSvcDeps, req requests.AcceptOrderRequest, mockErr error) {
	mockAcceptFailureBase(deps, req)
	deps.validator.ValidateAcceptMock.
		Expect(models.Order{}, req).
		Return(nil)
	deps.pricing.EvaluateMock.
		Expect(req.Package, req.Weight, req.Price).
		Return(75.0, nil)
	deps.actorSvc.DetermineActorMock.
		Expect(deps.ctx, models.EventAccepted, req.UserID).
		Return(models.Actor{}, nil)
	deps.repo.SaveMock.Set(func(ctx context.Context, order models.Order) error {
		return nil
	})
	deps.outboxRepo.CreateMock.Set(func(ctx context.Context, eventID uint64, orderID uint64, payload []byte) error {
		return nil
	})
	deps.history.RecordMock.Set(func(ctx context.Context, entry models.HistoryEntry) error {
		return mockErr
	})
}

func mockAcceptFailure(deps orderSvcDeps, req requests.AcceptOrderRequest, acceptanceStage acceptanceStage, mockErr error) {
	switch acceptanceStage {
	case stageValidate:
		mockAcceptFailureValidation(deps, req, mockErr)
	case stageEvaluate:
		mockAcceptFailureEvaluation(deps, req, mockErr)
	case stageSave:
		mockAcceptFailureSave(deps, req, mockErr)
	case stageOutbox:
		mockAcceptFailureOutbox(deps, req, mockErr)
	case stageRecord:
		mockAcceptFailureRecord(deps, req, mockErr)
	}
}
