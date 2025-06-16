package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// MapListReturnsParams converts CLI params for list returns command into internal request model
func (f *DefaultCLIFacadeMapper) MapListReturnsParams(p params.ListReturnsParams) (requests.OrdersFilterRequest, error) {
	if err := validatePaginationInfo(p.Page, p.Limit); err != nil {
		return requests.OrdersFilterRequest{}, err
	}

	status := models.Returned
	var opts []requests.FilterOption
	opts = append(opts, requests.WithStatus(status))

	if p.Page != nil {
		opts = append(opts, requests.WithPage(*p.Page))
	}
	if p.Limit != nil {
		opts = append(opts, requests.WithLimit(*p.Limit))
	}

	filter := requests.NewOrdersFilter(opts...)
	return filter, nil
}
