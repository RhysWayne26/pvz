package gateway

import (
	"log"
	"net"

	pb "pvz-cli/internal/gen/orders"

	"google.golang.org/grpc"
)

// RunGRPCServer starts the gRPC server on the given address. The provided impl must implement the pb.OrdersServiceServer interface.
func RunGRPCServer(grpcAddr string, impl pb.OrdersServiceServer) error {
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return err
	}

	srv := grpc.NewServer()
	pb.RegisterOrdersServiceServer(srv, impl)

	log.Printf("gRPC server started on %s", grpcAddr)
	return srv.Serve(lis)
}
