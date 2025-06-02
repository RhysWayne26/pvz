package apperrors

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
)
