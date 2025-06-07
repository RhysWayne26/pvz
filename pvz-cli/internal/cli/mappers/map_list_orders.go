package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
	"strconv"
	"strings"
)

// MapListOrdersParams converts CLI params into ListOrdersRequest.
func (f *DefaultCLIFacadeMapper) MapListOrdersParams(p params.ListOrdersParams) (requests.ListOrdersRequest, error) {
	userID, err := strconv.ParseUint(strings.TrimSpace(p.UserID), 10, 64)
	if err != nil {
		return requests.ListOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid user_id format")
	}

	if err := utils.ValidatePositiveInt("last", p.Last); err != nil {
		return requests.ListOrdersRequest{}, err
	}
	if err := utils.ValidatePositiveInt("page", p.Page); err != nil {
		return requests.ListOrdersRequest{}, err
	}
	if err := utils.ValidatePositiveInt("limit", p.Limit); err != nil {
		return requests.ListOrdersRequest{}, err
	}

	var lastID *uint64
	if raw := strings.TrimSpace(p.LastID); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return requests.ListOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid last_id format")
		}
		if id == 0 {
			return requests.ListOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "last_id must be > 0")
		}
		lastID = &id
	}

	return requests.ListOrdersRequest{
		UserID: userID,
		InPvz:  p.InPvz,
		LastID: lastID,
		Page:   p.Page,
		Limit:  p.Limit,
		Last:   p.Last,
	}, nil
}
