package interceptors

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var tracer = otel.Tracer("orders-grpc-server")

// TracingInterceptor is a gRPC unary interceptor that starts an OpenTelemetry span for each incoming request.
func TracingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(md))
		}

		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		span.SetAttributes(attribute.String("rpc.method", info.FullMethod))
		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
		return resp, err
	}
}
