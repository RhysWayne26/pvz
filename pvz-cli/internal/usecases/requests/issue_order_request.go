package requests

import (
	"github.com/google/uuid"
)

type IssueOrderRequest struct {
	OrderIDs []uuid.UUID
	UserID   uuid.UUID
}
