package handlers

import (
	"context"
	"errors"
	"pvz-cli/internal/metrics"
	svcmocks "pvz-cli/internal/usecases/services/mocks"
	"pvz-cli/pkg/cache"
	"testing"

	"github.com/stretchr/testify/require"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

var errListFail = errors.New("fail")

// TestDefaultFacadeHandler_HandleAcceptOrder_ContextCanceled tests HandleAcceptOrder with a canceled context, expecting context.Canceled error.
func TestDefaultFacadeHandler_HandleAcceptOrder_ContextCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	svc := svcmocks.NewOrderServiceMock(t)
	c := cache.NewNoopCache()
	m, _ := metrics.NewNoopHandlerMetrics()
	h := NewDefaultFacadeHandler(svc, nil, c, m)
	_, err := h.HandleAcceptOrder(ctx, requests.AcceptOrderRequest{OrderID: 5})
	require.ErrorIs(t, err, context.Canceled)
}

// TestDefaultFacadeHandler_HandleListOrders tests the behavior of HandleListOrders in various scenarios such as success and errors.
func TestDefaultFacadeHandler_HandleListOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cases := []struct {
		name      string
		setup     func(t *testing.T) FacadeHandler
		ctx       context.Context
		expectErr error
		wantResp  responses.ListOrdersResponse
	}{
		{
			name: "context canceled",
			ctx: func() context.Context {
				cctx, cancel := context.WithCancel(ctx)
				cancel()
				return cctx
			}(),
			setup: func(t *testing.T) FacadeHandler {
				svc := svcmocks.NewOrderServiceMock(t)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(svc, nil, c, m)
			},
			expectErr: context.Canceled,
		},
		{
			name: "service error",
			ctx:  ctx,
			setup: func(t *testing.T) FacadeHandler {
				svc := svcmocks.NewOrderServiceMock(t)
				svc.ListOrdersMock.Expect(ctx, requests.OrdersFilterRequest{}).Return(nil, 0, 0, errListFail)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(svc, nil, c, m)
			},
			expectErr: errListFail,
		},
		{
			name: "success",
			ctx:  ctx,
			setup: func(t *testing.T) FacadeHandler {
				svc := svcmocks.NewOrderServiceMock(t)
				svc.ListOrdersMock.Expect(ctx, requests.OrdersFilterRequest{}).
					Return([]models.Order{{OrderID: 10}}, 10, 1, nil)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(svc, nil, c, m)
			},
			wantResp: responses.ListOrdersResponse{
				Orders: []models.Order{{OrderID: 10}},
				NextID: func() *uint64 { v := uint64(10); return &v }(),
				Total:  func() *int { v := 1; return &v }(),
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := tt.setup(t)
			resp, err := h.HandleListOrders(tt.ctx, requests.OrdersFilterRequest{})
			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

// TestDefaultFacadeHandler_HandleOrderHistory tests the handling of order history requests in DefaultFacadeHandler, ensuring proper behavior.
func TestDefaultFacadeHandler_HandleOrderHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cases := []struct {
		name      string
		setup     func(t *testing.T) FacadeHandler
		ctx       context.Context
		expectErr error
		wantResp  responses.OrderHistoryResponse
	}{
		{
			name: "context canceled",
			ctx: func() context.Context {
				cctx, cancel := context.WithCancel(ctx)
				cancel()
				return cctx
			}(),
			setup: func(t *testing.T) FacadeHandler {
				svc := svcmocks.NewOrderServiceMock(t)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(svc, nil, c, m)
			},
			expectErr: context.Canceled,
		},
		{
			name: "service error",
			ctx:  ctx,
			setup: func(t *testing.T) FacadeHandler {
				hsvc := svcmocks.NewHistoryServiceMock(t)
				hsvc.ListMock.Expect(ctx, requests.OrderHistoryFilter{}).Return(nil, errListFail)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(nil, hsvc, c, m)
			},
			expectErr: errListFail,
		},
		{
			name: "success",
			ctx:  ctx,
			setup: func(t *testing.T) FacadeHandler {
				hsvc := svcmocks.NewHistoryServiceMock(t)
				hsvc.ListMock.Expect(ctx, requests.OrderHistoryFilter{}).Return([]models.HistoryEntry{{OrderID: 5}}, nil)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(nil, hsvc, c, m)
			},
			wantResp: responses.OrderHistoryResponse{History: []models.HistoryEntry{{OrderID: 5}}},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := tt.setup(t)
			resp, err := h.HandleOrderHistory(tt.ctx, requests.OrderHistoryFilter{})
			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

// TestDefaultFacadeHandler_HandleReturnOrder tests the HandleReturnOrder method of DefaultFacadeHandler with various scenarios.
func TestDefaultFacadeHandler_HandleReturnOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		name      string
		setup     func(t *testing.T) FacadeHandler
		ctx       context.Context
		req       requests.ReturnOrderRequest
		expectErr error
		wantResp  responses.ReturnOrderResponse
	}{
		{
			name: "context canceled",
			ctx: func() context.Context {
				cctx, cancel := context.WithCancel(ctx)
				cancel()
				return cctx
			}(),
			req: requests.ReturnOrderRequest{},
			setup: func(t *testing.T) FacadeHandler {
				svc := svcmocks.NewOrderServiceMock(t)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(svc, nil, c, m)
			},
			expectErr: context.Canceled,
		},
		{
			name: "service error",
			ctx:  ctx,
			req:  requests.ReturnOrderRequest{OrderID: 8},
			setup: func(t *testing.T) FacadeHandler {
				svc := svcmocks.NewOrderServiceMock(t)
				svc.ReturnToCourierMock.Expect(ctx, requests.ReturnOrderRequest{OrderID: 8}).Return(errListFail)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(svc, nil, c, m)
			},
			expectErr: errListFail,
		},
		{
			name: "success",
			ctx:  ctx,
			req:  requests.ReturnOrderRequest{OrderID: 8},
			setup: func(t *testing.T) FacadeHandler {
				svc := svcmocks.NewOrderServiceMock(t)
				svc.ReturnToCourierMock.Expect(ctx, requests.ReturnOrderRequest{OrderID: 8}).Return(nil)
				c := cache.NewNoopCache()
				m, _ := metrics.NewNoopHandlerMetrics()
				return NewDefaultFacadeHandler(svc, nil, c, m)
			},
			wantResp: responses.ReturnOrderResponse{OrderID: 8},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := tt.setup(t)
			resp, err := h.HandleReturnOrder(tt.ctx, tt.req)
			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

// TestDefaultFacadeHandler_HandleImportOrders tests the order import handling logic in the DefaultFacadeHandler; the test verifies successful imports and error propagation for failed imports.
func TestDefaultFacadeHandler_HandleImportOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	statuses := []requests.ImportOrderStatus{
		{ItemNumber: 1, Request: &requests.AcceptOrderRequest{OrderID: 1}},
		{ItemNumber: 2, Request: &requests.AcceptOrderRequest{OrderID: 2}},
		{ItemNumber: 3, Request: &requests.AcceptOrderRequest{OrderID: 3}, Error: errors.New("already invalid")},
	}
	svc := svcmocks.NewOrderServiceMock(t)
	defer svc.MinimockFinish()
	svc.ImportOrdersMock.
		Expect(ctx, requests.ImportOrdersRequest{Statuses: statuses}).
		Return([]models.BatchEntryProcessedResult{
			{OrderID: 1, Error: nil},
			{OrderID: 2, Error: errors.New("fail2")},
			{OrderID: 3, Error: errors.New("already invalid")},
		}, nil)
	c := cache.NewNoopCache()
	m, _ := metrics.NewNoopHandlerMetrics()
	h := NewDefaultFacadeHandler(svc, nil, c, m)
	resp, err := h.HandleImportOrders(ctx, requests.ImportOrdersRequest{Statuses: statuses})
	require.NoError(t, err)
	require.Equal(t, 1, resp.Imported)
	require.NoError(t, resp.Statuses[0].Error)
	require.EqualError(t, resp.Statuses[1].Error, "fail2")
	require.EqualError(t, resp.Statuses[2].Error, "already invalid")
}

// TestDefaultFacadeHandler_HandleProcessOrders tests the HandleProcessOrders function ensuring actions like issue and return work correctly.
func TestDefaultFacadeHandler_HandleProcessOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Run("issue and return branch", func(t *testing.T) {
		t.Parallel()
		svc := svcmocks.NewOrderServiceMock(t)
		defer svc.MinimockFinish()
		svc.IssueOrdersMock.Expect(ctx, requests.IssueOrdersRequest{UserID: 1, OrderIDs: []uint64{10}}).
			Return([]models.BatchEntryProcessedResult{{OrderID: 10}}, nil)
		svc.CreateClientReturnsMock.Expect(ctx, requests.ClientReturnsRequest{UserID: 1, OrderIDs: []uint64{11}}).
			Return([]models.BatchEntryProcessedResult{{OrderID: 11}}, nil)
		c := cache.NewNoopCache()
		m, _ := metrics.NewNoopHandlerMetrics()
		h := NewDefaultFacadeHandler(svc, nil, c, m)
		resp1, err1 := h.HandleProcessOrders(ctx, requests.ProcessOrdersRequest{
			UserID:   1,
			OrderIDs: []uint64{10},
			Action:   requests.ActionIssue,
		})
		require.NoError(t, err1)
		require.ElementsMatch(t, []uint64{10}, resp1.Processed)
		resp2, err2 := h.HandleProcessOrders(ctx, requests.ProcessOrdersRequest{
			UserID:   1,
			OrderIDs: []uint64{11},
			Action:   requests.ActionReturn,
		})
		require.NoError(t, err2)
		require.ElementsMatch(t, []uint64{11}, resp2.Processed)
	})
}
