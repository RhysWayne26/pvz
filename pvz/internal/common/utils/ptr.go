package utils

// Ptr returns pointer for a chosen value
func Ptr[T any](v T) *T {
	return &v
}
