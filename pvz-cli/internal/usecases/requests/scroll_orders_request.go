package requests

import "github.com/google/uuid"

type ScrollOrdersRequest struct {
	UserID  uuid.UUID
	Limit   *int
	AfterID *uuid.UUID
}
