package apperrors

type ErrorCode string

const (
	OrderNotFound      ErrorCode = "ORDER_NOT_FOUND"
	OrderAlreadyExists ErrorCode = "ORDER_ALREADY_EXISTS"
	StorageExpired     ErrorCode = "STORAGE_EXPIRED"
	ValidationFailed   ErrorCode = "VALIDATION_FAILED"
	InternalError      ErrorCode = "INTERNAL_ERROR"
)
