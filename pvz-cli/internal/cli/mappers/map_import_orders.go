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

	var statuses []requests.ImportOrderStatus
	for i, raw := range rawOrders {
		itemNumber := i + 1
		status := requests.ImportOrderStatus{ItemNumber: itemNumber}
		acceptRequest, err := f.MapAcceptOrderParams(raw)
		if err != nil {
			status.Error = apperrors.Newf(apperrors.InvalidBatchEntry, "order #%d: %v", itemNumber, err)
		} else {
			status.Request = &acceptRequest
		}
		statuses = append(statuses, status)
	}
	return requests.ImportOrdersRequest{
		Statuses: statuses,
	}, nil
}
