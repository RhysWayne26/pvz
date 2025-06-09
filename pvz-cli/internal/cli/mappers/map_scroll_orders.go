package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
	"strconv"
	"strings"
)

// MapScrollOrdersParams converts CLI params for scroll orders command into internal request model
func (f *DefaultCLIFacadeMapper) MapScrollOrdersParams(p params.ScrollOrdersParams) (requests.OrdersFilterRequest, error) {
	userID, err := strconv.ParseUint(strings.TrimSpace(p.UserID), 10, 64)
	if err != nil {
		return requests.OrdersFilterRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid user_id format")
	}

	if err := utils.ValidatePositiveInt("limit", p.Limit); err != nil {
		return requests.OrdersFilterRequest{}, err
	}

	var lastID *uint64
	if strings.TrimSpace(p.LastID) != "" {
		id, err := strconv.ParseUint(strings.TrimSpace(p.LastID), 10, 64)
		if err != nil {
			return requests.OrdersFilterRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid last_id format")
		}
		lastID = &id
	}

	return requests.OrdersFilterRequest{
		UserID: &userID,
		Limit:  p.Limit,
		LastID: lastID,
	}, nil
}
