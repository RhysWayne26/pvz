package apperrors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func Newf(code ErrorCode, format string, args ...any) error {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func Handle(err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		fmt.Printf("ERROR: %s: %s\n", appErr.Code, appErr.Message)
	} else {
		fmt.Printf("ERROR: INTERNAL_ERROR: %v\n", err)
	}
}
