package interceptors

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const correlationIDHeader = "x-correlation-id"

type correlationIDCtxKey struct{}

var correlationIDKey = correlationIDCtxKey{}

// CorrelationIDInterceptor extracts or injects X-Correlation-ID into context and metadata.
func CorrelationIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		var corrID string
		if vals := md[correlationIDHeader]; len(vals) > 0 && vals[0] != "" {
			corrID = vals[0]
		} else {
			corrID = uuid.NewString()
			md.Set(correlationIDHeader, corrID)
			ctx = metadata.NewIncomingContext(ctx, md)
		}
		ctx = context.WithValue(ctx, correlationIDKey, corrID)
		return handler(ctx, req)
	}
}
