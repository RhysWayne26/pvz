package apperrors

import (
	"errors"

	"google.golang.org/grpc/status"
)

// ErrorCode represents specific application error codes
type ErrorCode string

// Application error codes for different failure scenarios
const (
	OrderNotFound      ErrorCode = "ORDER_NOT_FOUND"
	OrderAlreadyExists ErrorCode = "ORDER_ALREADY_EXISTS"
	StorageExpired     ErrorCode = "STORAGE_EXPIRED"
	ValidationFailed   ErrorCode = "VALIDATION_FAILED"
	InternalError      ErrorCode = "INTERNAL_ERROR"
	InvalidPackage     ErrorCode = "INVALID_PACKAGE"
	WeightTooHeavy     ErrorCode = "WEIGHT_TOO_HEAVY"
	InvalidBatchEntry  ErrorCode = "INVALID_BATCH_ENTRY"
	InvalidID          ErrorCode = "INVALID_ID"
)

// CodeFromError helps to extract code from application error common struct
func CodeFromError(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return string(appErr.Code)
	}

	if st, ok := status.FromError(err); ok {
		return st.Code().String()
	}
	return string(InternalError)
}
