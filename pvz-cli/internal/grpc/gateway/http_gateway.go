package gateway

import (
	"context"
	"log"
	"net/http"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "pvz-cli/internal/gen/orders"
)

// RunHTTPGateway starts the HTTP reverse-proxy server for gRPC.
// It connects to the running gRPC server at grpcAddr and exposes the HTTP API at httpAddr.
func RunHTTPGateway(grpcAddr string, httpAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := pb.RegisterOrdersServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	if err != nil {
		return err
	}

	log.Printf("HTTP gateway started on %s", httpAddr)
	return http.ListenAndServe(httpAddr, mux)
}
