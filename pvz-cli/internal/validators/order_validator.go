package validators

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// OrderValidator validates order operations and business rules
type OrderValidator interface {
	ValidateAccept(o models.Order, req requests.AcceptOrderRequest) error
	ValidateIssue(orders []models.Order, req requests.IssueOrdersRequest) error
}
