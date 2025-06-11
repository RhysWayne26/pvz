package interceptors

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

// LoggingInterceptor returns a grpc.UnaryServerInterceptor that logs each gRPC call.
// It records method, status, latency, correlation ID (corr_id), trace ID (trace_id), request and response payloads using a structured zap.Logger.
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(start)
		st, _ := status.FromError(err)
		code := st.Code().String()
		corrID, _ := ctx.Value(correlationIDKey).(string)
		spanCtx := trace.SpanContextFromContext(ctx)
		traceID := ""
		if spanCtx.IsValid() {
			traceID = spanCtx.TraceID().String()
		}

		logger.Info("gRPC call",
			zap.String("method", info.FullMethod),
			zap.String("status", code),
			zap.Duration("latency", elapsed),
			zap.String("corr_id", corrID),
			zap.String("trace_id", traceID),
			zap.Any("request", req),
			zap.Any("response", resp),
		)
		return resp, err
	}
}
