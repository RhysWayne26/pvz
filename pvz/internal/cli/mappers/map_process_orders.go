package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/usecases/requests"
	"strconv"
	"strings"
)

// MapProcessOrdersParams converts CLI params for process orders command into internal request model
func (f *DefaultCLIFacadeMapper) MapProcessOrdersParams(p params.ProcessOrdersParams) (requests.ProcessOrdersRequest, error) {
	userID, err := strconv.ParseUint(strings.TrimSpace(p.UserID), 10, 64)
	if err != nil {
		return requests.ProcessOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid user_id format")
	}

	rawIDs := strings.Split(p.OrderIDs, ",")
	if len(rawIDs) == 0 || (len(rawIDs) == 1 && strings.TrimSpace(rawIDs[0]) == "") {
		return requests.ProcessOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "no order IDs provided")
	}

	parsedIDs := make([]uint64, 0, len(rawIDs))
	for i, raw := range rawIDs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return requests.ProcessOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "empty order ID at position %d", i+1)
		}
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return requests.ProcessOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "invalid order ID %q at position %d", raw, i+1)
		}
		parsedIDs = append(parsedIDs, id)
	}

	action := strings.TrimSpace(p.Action)
	switch action {
	case string(requests.ActionIssue), string(requests.ActionReturn):
		return requests.ProcessOrdersRequest{
			UserID:   userID,
			OrderIDs: parsedIDs,
			Action:   requests.ProcessAction(action),
		}, nil
	default:
		return requests.ProcessOrdersRequest{}, apperrors.Newf(apperrors.ValidationFailed, "unknown action %q", action)
	}
}
