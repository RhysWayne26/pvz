package gateway

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
	adminpb "pvz-cli/internal/gen/admin"
	pb "pvz-cli/internal/gen/orders"
	"time"
)

// RunHTTPGateway starts the HTTP <-> gRPC reverse-proxy. It connects to the gRPC server at grpcAddr and serves HTTP on httpAddr.
func RunHTTPGateway(ctx context.Context, ordersGrpcAddr, adminGrpcAddr, httpAddr string) error {
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
		ordersGrpcAddr,
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
	if err := adminpb.RegisterAdminServiceHandlerFromEndpoint(
		ctx,
		mux,
		adminGrpcAddr,
		[]grpc.DialOption{
			grpc.WithContextDialer(dialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	); err != nil {
		return fmt.Errorf("failed to register admin gateway handler: %w", err)
	}

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(mux)

	srv := &http.Server{
		Addr:              httpAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

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
