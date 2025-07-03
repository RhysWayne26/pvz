//go:build e2e

package standalone

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net"
	"os"
	"pvz-cli/internal/app"
	"sync"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/tests"
)

func TestE2E_AcceptIssueAndReturn(t *testing.T) {
	r := runner.NewRunner(t, "E2E: Accept to Issue to Return")
	r.NewTest("Accept order", func(t provider.T) {
		t.Parallel()
		const (
			orderID uint64 = 456
			userID  uint64 = 333
		)
		deps := newE2E(t)
		t.WithNewStep(
			fmt.Sprintf("Accept order #%d by user #%d", orderID, userID),
			func(ctx provider.StepCtx) {
				req := &pb.AcceptOrderRequest{
					OrderId:   orderID,
					UserId:    userID,
					ExpiresAt: timestamppb.New(time.Now().Add(2 * time.Hour)),
					Package:   pb.PackageType_PACKAGE_TYPE_BOX.Enum(),
					Weight:    1.5,
					Price:     50,
				}
				res, err := deps.client.AcceptOrder(context.Background(), req)
				require.NoError(t, err)
				require.Equal(t, orderID, res.OrderId)
				require.Equal(t, pb.OrderStatus_ORDER_STATUS_ACCEPTED, res.Status)
			},
		)
	})
	r.NewTest("Issue order", func(t provider.T) {
		t.Parallel()
		const (
			orderID uint64 = 457
			userID  uint64 = 334
		)
		deps := newE2E(t)
		t.WithNewStep(
			fmt.Sprintf("Setup: accept order #%d", orderID),
			func(ctx provider.StepCtx) {
				_, err := deps.client.AcceptOrder(context.Background(), &pb.AcceptOrderRequest{
					OrderId:   orderID,
					UserId:    userID,
					ExpiresAt: timestamppb.New(time.Now().Add(2 * time.Hour)),
					Package:   pb.PackageType_PACKAGE_TYPE_BOX.Enum(),
					Weight:    1.5,
					Price:     50,
				})
				require.NoError(t, err)
			},
		)
		t.WithNewStep(
			fmt.Sprintf("Issue order #%d by user #%d", orderID, userID),
			func(ctx provider.StepCtx) {
				issueReq := &pb.ProcessOrdersRequest{
					UserId:   userID,
					OrderIds: []uint64{orderID},
					Action:   pb.ActionType_ACTION_TYPE_ISSUE,
				}

				issueRes, err := deps.client.ProcessOrders(context.Background(), issueReq)
				require.NoError(t, err)
				require.Contains(t, issueRes.Processed, orderID)
				require.Empty(t, issueRes.Errors)
			},
		)
	})

	r.NewTest("Return order", func(t provider.T) {
		t.Parallel()
		cases := []struct {
			orderID uint64
			userID  uint64
		}{
			{orderID: 1001, userID: 7},
			{orderID: 2002, userID: 8},
		}
		deps := newE2E(t)
		for _, tc := range cases {
			tc := tc
			t.WithNewStep(
				fmt.Sprintf("Setup: accept & issue order #%d by user #%d", tc.orderID, tc.userID),
				func(ctx provider.StepCtx) {
					_, err := deps.client.AcceptOrder(context.Background(), &pb.AcceptOrderRequest{
						OrderId:   tc.orderID,
						UserId:    tc.userID,
						ExpiresAt: timestamppb.New(time.Now().Add(2 * time.Hour)),
						Package:   pb.PackageType_PACKAGE_TYPE_BOX.Enum(),
						Weight:    2.0,
						Price:     75,
					})
					require.NoError(t, err)
					_, err = deps.client.ProcessOrders(context.Background(), &pb.ProcessOrdersRequest{
						UserId:   tc.userID,
						OrderIds: []uint64{tc.orderID},
						Action:   pb.ActionType_ACTION_TYPE_ISSUE,
					})
					require.NoError(t, err)
				},
			)
			t.WithNewStep(
				fmt.Sprintf("Return order #%d by user #%d", tc.orderID, tc.userID),
				func(ctx provider.StepCtx) {
					returnRes, err := deps.client.ProcessOrders(context.Background(), &pb.ProcessOrdersRequest{
						UserId:   tc.userID,
						OrderIds: []uint64{tc.orderID},
						Action:   pb.ActionType_ACTION_TYPE_RETURN,
					})
					require.NoError(t, err)
					require.Contains(t, returnRes.Processed, tc.orderID)
					require.Empty(t, returnRes.Errors)
				},
			)
		}
	})
	r.RunTests()
}

func findFreePort(t provider.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	defer func(l *net.TCPListener) {
		err := l.Close()
		if err != nil {
			slog.Warn("Failed to close TCP listener", "error", err)
		}
	}(l)
	return l.Addr().(*net.TCPAddr).Port
}

func newE2E(t provider.T) e2eDeps {
	commonDeps := tests.NewCommonDeps(t)
	err := os.Setenv("APP_ENV", "test")
	if err != nil {
		slog.Warn("Failed to set APP_ENV: ", "error", err)
	}
	err = os.Setenv("TEST_DB_DSN", commonDeps.GetDSN())
	if err != nil {
		slog.Warn("Failed to set TEST_DB_DSN: ", "error", err)
	}
	t.Cleanup(func() {
		err := os.Setenv("APP_ENV", "")
		if err != nil {
			slog.Warn("Failed to unset APP_ENV", "error", err)
		}
		err = os.Setenv("TEST_DB_DSN", "")
		if err != nil {
			slog.Warn("Failed to unset TEST_DB_DSN", "error", err)
		}
	})
	application := app.New()
	t.Cleanup(func() {
		application.Shutdown()
	})
	port := findFreePort(t)
	portStr := fmt.Sprintf(":%d", port)
	var wg sync.WaitGroup
	wg.Add(1)
	go app.StartGRPCServer(application, portStr, &wg)
	time.Sleep(5 * time.Second)
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() {
		conn.Close()
	})
	client := pb.NewOrdersServiceClient(conn)
	return e2eDeps{client: client}
}

type e2eDeps struct {
	client pb.OrdersServiceClient
}
