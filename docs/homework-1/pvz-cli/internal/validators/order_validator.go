package validators

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

type OrderValidator interface {
	ValidateAccept(o models.Order, req requests.AcceptOrderRequest) error
	ValidateIssue(orders []models.Order, req requests.IssueOrderRequest) error
}
