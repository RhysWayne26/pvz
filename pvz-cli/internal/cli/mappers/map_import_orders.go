package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
)

// MapImportOrdersParams parses and maps file data into ImportOrdersRequest
func (f *DefaultCLIFacadeMapper) MapImportOrdersParams(p params.ImportOrdersParams) (requests.ImportOrdersRequest, error) {
	if p.File == "" {
		return requests.ImportOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "file path must not be empty")
	}

	rawOrders, err := utils.ParseOrdersFromFile(p.File)
	if err != nil {
		return requests.ImportOrdersRequest{}, err
	}

	var result []requests.AcceptOrderRequest
	for i, raw := range rawOrders {
		order, err := f.MapAcceptOrderParams(raw)
		if err != nil {
			return requests.ImportOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "order #%d: %v", i+1, err)
		}
		result = append(result, order)
	}

	return requests.ImportOrdersRequest{
		Orders: result,
	}, nil
}
