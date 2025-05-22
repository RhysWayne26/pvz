package requests

import (
	"github.com/google/uuid"
)

type ReturnOrderRequest struct {
	OrderID uuid.UUID
}
