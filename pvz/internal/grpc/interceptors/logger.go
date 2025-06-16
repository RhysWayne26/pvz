package interceptors

import (
	"context"
	"log"
	"os"
	"pvz-cli/internal/common/apperrors"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	logger  *zap.Logger
	logFile *os.File
)

func init() {
	err := os.MkdirAll("logs", 0750)
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.InternalError, "failed to create log directory: %v", err))
	}

	logFile, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.InternalError, "failed to create log file: %v", err))
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(logFile),
		zap.InfoLevel,
	)
	logger = zap.New(core, zap.AddCaller())
	log.SetOutput(logFile)
}

// GetLogger returns zap logger pointer
func GetLogger() *zap.Logger {
	return logger
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

// CloseLogFile for safely closing log file
func CloseLogFile() error {
	if logFile != nil {
		err := logFile.Close()
		if err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to close log file: %v", err)
		}
	}
	return nil
}
