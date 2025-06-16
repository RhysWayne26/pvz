package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"os"
)

var tracerProvider *sdktrace.TracerProvider

// InitTracing sets up OpenTelemetry tracing with a stdout exporter.
func InitTracing(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	file, err := os.OpenFile("logs/traces.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	exp, err := stdouttrace.New(
		stdouttrace.WithWriter(file),
		stdouttrace.WithoutTimestamps(),
	)
	if err != nil {
		return err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("pvz"),
		)),
	)
	otel.SetTracerProvider(tp)
	tracerProvider = tp
	return nil
}

// ShutdownTracing flushes and shuts down the tracer provider if initialized.
func ShutdownTracing(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if tracerProvider != nil {
		return tracerProvider.Shutdown(ctx)
	}
	return nil
}
