package requests

import "github.com/google/uuid"

type ListOrdersFilter struct {
	UserID uuid.UUID
	InPvz  *bool
	LastID *uuid.UUID
	Page   *int
	Limit  *int
	Last   *int
}
