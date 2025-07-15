package gateway

import (
	"context"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"

	pb "pvz-cli/internal/gen/orders"

	"google.golang.org/grpc"
)

// RunGRPCServer starts the gRPC server on the given address. The provided impl must implement the pb.OrdersServiceServer interface.
func RunGRPCServer(ctx context.Context, grpcAddr string, svc pb.OrdersServiceServer, opts ...grpc.ServerOption) error {
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterOrdersServiceServer(srv, svc)
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("orders.OrdersService", grpc_health_v1.HealthCheckResponse_SERVING)
	log.Printf("gRPC server started on %s", grpcAddr)
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	return srv.Serve(lis)
}
