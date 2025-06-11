package gateway

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "pvz-cli/internal/gen/orders"
)

// RunHTTPGateway starts the HTTP reverse-proxy server for gRPC.
// It connects to the running gRPC server at grpcAddr and exposes the HTTP API at httpAddr.
func RunHTTPGateway(grpcAddr, httpAddr string) error {
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC client: %w", err)
	}

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(GRPCGatewayErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
				UseEnumNumbers:  false,
			},
		}),
	)

	if err := pb.RegisterOrdersServiceHandler(
		context.Background(),
		mux,
		conn,
	); err != nil {
		return fmt.Errorf("failed to register gateway handler: %w", err)
	}

	return http.ListenAndServe(httpAddr, mux)
}
