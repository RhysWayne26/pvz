package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const defaultJaegerEndpoint = "http://localhost:14268/api/traces"

var tracerProvider *sdktrace.TracerProvider

// InitTracing sets up OpenTelemetry tracing, exporting to Jaeger. It first checks JAEGER_ENDPOINT (for local runs), then JAEGER_COLLECTOR_PORT (for Docker Compose).
// If neither is set, it falls back to defaultJaegerEndpoint.
func InitTracing(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if ep := os.Getenv("JAEGER_ENDPOINT"); ep != "" {
		slog.Info("Initializing Jaeger exporter using JAEGER_ENDPOINT", "endpoint", ep)
		return initJaeger(ep)
	}
	portMap := os.Getenv("JAEGER_COLLECTOR_PORT")
	if portMap != "" {
		parts := strings.Split(portMap, ":")
		if len(parts) != 2 {
			return fmt.Errorf("JAEGER_COLLECTOR_PORT=%q invalid, must be host:container", portMap)
		}
		endpoint := fmt.Sprintf("http://jaeger:%s/api/traces", parts[1])
		slog.Info("Initializing Jaeger exporter using JAEGER_COLLECTOR_PORT", "endpoint", endpoint)
		return initJaeger(endpoint)
	}
	slog.Info("Initializing Jaeger exporter using default endpoint", "endpoint", defaultJaegerEndpoint)
	return initJaeger(defaultJaegerEndpoint)
}

func initJaeger(endpoint string) error {
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)),
	)
	if err != nil {
		return fmt.Errorf("failed to create jaeger exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
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
