package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
	"strconv"
	"strings"
)

// MapListOrdersParams converts CLI params into OrdersFilterRequest.
func (f *DefaultCLIFacadeMapper) MapListOrdersParams(p params.ListOrdersParams) (requests.OrdersFilterRequest, error) {
	userID, err := strconv.ParseUint(strings.TrimSpace(p.UserID), 10, 64)
	if err != nil {
		return requests.OrdersFilterRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid user_id format")
	}

	if err := utils.ValidatePositiveInt("last", p.Last); err != nil {
		return requests.OrdersFilterRequest{}, err
	}
	if err := utils.ValidatePositiveInt("page", p.Page); err != nil {
		return requests.OrdersFilterRequest{}, err
	}
	if err := utils.ValidatePositiveInt("limit", p.Limit); err != nil {
		return requests.OrdersFilterRequest{}, err
	}

	var lastIDVal uint64
	var lastID *uint64
	if raw := strings.TrimSpace(p.LastID); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return requests.OrdersFilterRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid last_id format")
		}
		if id == 0 {
			return requests.OrdersFilterRequest{}, apperrors.Newf(apperrors.ValidationFailed, "last_id must be > 0")
		}
		lastIDVal = id
		lastID = &lastIDVal
	}

	var opts []requests.FilterOption
	opts = append(opts, requests.WithUserID(userID))
	if p.InPvz != nil {
		opts = append(opts, requests.WithInPvz(*p.InPvz))
	}
	if lastID != nil {
		opts = append(opts, requests.WithLastID(*lastID))
	}
	if p.Page != nil {
		opts = append(opts, requests.WithPage(*p.Page))
	}
	if p.Limit != nil {
		opts = append(opts, requests.WithLimit(*p.Limit))
	}
	if p.Last != nil {
		opts = append(opts, requests.WithLast(*p.Last))
	}

	filter := requests.NewOrdersFilter(opts...)
	return filter, nil
}
