package dto

// ListReturnsParams contains parameters for list-returns command
type ListReturnsParams struct {
	Page  *int `json:"page,omitempty"`
	Limit *int `json:"limit,omitempty"`
}
