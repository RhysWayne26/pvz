package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
)

// ValidationInterceptor checks for a Validate() error method on the request. If present and returns an error, the request is rejected with InvalidArgument.
func ValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		val := reflect.ValueOf(req)
		method := val.MethodByName("Validate")
		if method.IsValid() {
			results := method.Call(nil)
			if len(results) == 1 && !results[0].IsNil() {
				err := results[0].Interface().(error)
				return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
			}
		}
		return handler(ctx, req)
	}
}
