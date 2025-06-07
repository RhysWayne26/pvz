package apperrors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// GatewayErrorHandler is a custom error handler for grpc-gateway. It maps internal AppError and gRPC errors to a consistent JSON HTTP response.
func GatewayErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error) {

	w.Header().Set("Content-Type", "application/json")
	var appErr *AppError
	var (
		code       string
		message    string
		httpStatus int
	)

	if errors.As(err, &appErr) {
		code = string(appErr.Code)
		message = appErr.Message
		httpStatus = http.StatusBadRequest
	} else {
		st, _ := status.FromError(err)
		code = "INTERNAL_ERROR"
		message = st.Message()
		httpStatus = http.StatusInternalServerError
	}

	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
