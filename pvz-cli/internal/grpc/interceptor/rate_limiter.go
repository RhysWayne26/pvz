package interceptor

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	once    sync.Once
	limiter *rate.Limiter
)

// RateLimitInterceptor returns a UnaryServerInterceptor that allows up to 5 RPS.
func RateLimitInterceptor() grpc.UnaryServerInterceptor {
	once.Do(func() {
		limiter = rate.NewLimiter(rate.Every(time.Second/5), 1)
	})

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !limiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "RATE_LIMITED: too many requests")
		}
		return handler(ctx, req)
	}
}
