package gateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	pb "pvz-cli/internal/gen/orders"
)

// RunHTTPGateway starts the HTTP <-> gRPC reverse-proxy. It connects to the gRPC server at grpcAddr and serves HTTP on httpAddr.
func RunHTTPGateway(ctx context.Context, grpcAddr, httpAddr string) error {
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
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
	if err := pb.RegisterOrdersServiceHandlerFromEndpoint(
		ctx,
		mux,
		grpcAddr,
		[]grpc.DialOption{
			grpc.WithContextDialer(dialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithConnectParams(grpc.ConnectParams{
				MinConnectTimeout: 5 * time.Second,
			}),
		},
	); err != nil {
		return fmt.Errorf("failed to register grpc gateway handler: %w", err)
	}
	srv := &http.Server{Addr: httpAddr, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP gateway shutdown error: %v", err)
		}
	}()

	return srv.ListenAndServe()
}
