package requests

import "github.com/google/uuid"

type ClientReturnRequest struct {
	OrderIDs []uuid.UUID
	UserID   uuid.UUID
}
