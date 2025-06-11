package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"pvz-cli/internal/common/apperrors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCGatewayErrorHandler maps gRPC and internal errors to consistent HTTP JSON error responses.
func GRPCGatewayErrorHandler(
	_ context.Context, _ *runtime.ServeMux, _ runtime.Marshaler,
	w http.ResponseWriter, _ *http.Request, err error,
) {
	w.Header().Set("Content-Type", "application/json")

	var (
		code       string
		message    string
		httpStatus int
	)

	if err == nil {
		code = "UNKNOWN"
		message = "no error provided"
		httpStatus = http.StatusInternalServerError
	} else if appErr := new(apperrors.AppError); errors.As(err, &appErr) {
		code = string(appErr.Code)
		message = appErr.Message
		switch appErr.Code {
		case apperrors.OrderAlreadyExists:
			httpStatus = http.StatusConflict
		case apperrors.OrderNotFound:
			httpStatus = http.StatusNotFound
		case apperrors.StorageExpired,
			apperrors.WeightTooHeavy:
			httpStatus = http.StatusPreconditionFailed
		default:
			httpStatus = http.StatusBadRequest
		}

	} else {
		st, _ := status.FromError(err)
		code = st.Code().String()
		message = st.Message()
		httpStatus = grpcCodeToHTTPStatus(st.Code())
	}

	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    code,
		"message": message,
	})
}

func grpcCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
