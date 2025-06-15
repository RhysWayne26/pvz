package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/usecases/requests"
	"strconv"
	"strings"
)

// MapReturnOrderParams converts CLI params for return order command into internal request model
func (f *DefaultCLIFacadeMapper) MapReturnOrderParams(p params.ReturnOrderParams) (requests.ReturnOrderRequest, error) {
	orderID, err := strconv.ParseUint(strings.TrimSpace(p.OrderID), 10, 64)
	if err != nil {
		return requests.ReturnOrderRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid order_id format")
	}

	return requests.ReturnOrderRequest{
		OrderID: orderID,
	}, nil
}
