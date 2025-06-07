package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
)

// MapListReturnsParams converts CLI params for list returns command into internal request model
func (f *DefaultCLIFacadeMapper) MapListReturnsParams(p params.ListReturnsParams) (requests.ListReturnsRequest, error) {
	if err := utils.ValidatePositiveInt("page", p.Page); err != nil {
		return requests.ListReturnsRequest{}, err
	}
	if err := utils.ValidatePositiveInt("limit", p.Limit); err != nil {
		return requests.ListReturnsRequest{}, err
	}

	page := constants.DefaultPage
	limit := constants.DefaultLimit

	if p.Page != nil {
		page = *p.Page
	}
	if p.Limit != nil {
		limit = *p.Limit
	}

	return requests.ListReturnsRequest{
		Page:  page,
		Limit: limit,
	}, nil
}
