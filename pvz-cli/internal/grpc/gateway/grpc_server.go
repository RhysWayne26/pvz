package gateway

import (
	"context"
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
	log.Printf("gRPC server started on %s", grpcAddr)
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	return srv.Serve(lis)
}
