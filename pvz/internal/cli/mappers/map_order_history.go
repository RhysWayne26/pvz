package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/usecases/requests"
)

// MapOrderHistoryParams converts CLI params for order-history command into internal request model
func (f *DefaultCLIFacadeMapper) MapOrderHistoryParams(p params.OrderHistoryParams) (requests.OrderHistoryRequest, error) {
	if err := validatePaginationInfo(p.Page, p.Limit); err != nil {
		return requests.OrderHistoryRequest{}, err
	}
	req := requests.OrderHistoryRequest{
		Page:  constants.DefaultHistoryPage,
		Limit: constants.DefaultHistoryLimit,
	}
	if p.Page != nil {
		req.Page = *(p.Page)
	}
	if p.Limit != nil {
		req.Limit = *(p.Limit)
	}

	return req, nil
}
