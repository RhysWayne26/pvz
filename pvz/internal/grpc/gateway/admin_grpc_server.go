package gateway

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"pvz-cli/internal/gen/admin"
)

// RunAdminGRPCServer starts the Admin gRPC server on the specified address with the provided service and options.
func RunAdminGRPCServer(ctx context.Context, grpcAddr string, svc admin.AdminServiceServer, opts ...grpc.ServerOption) error {
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer(opts...)
	admin.RegisterAdminServiceServer(srv, svc)
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("admin.AdminService", grpc_health_v1.HealthCheckResponse_SERVING)
	log.Printf("Admin gRPC server started on %s", grpcAddr)
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	return srv.Serve(lis)
}
