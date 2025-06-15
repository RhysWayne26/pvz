package mappers

import (
	"pvz-cli/internal/common/utils"
)

func validatePaginationInfo(page *int, limit *int) error {
	if err := utils.ValidatePositiveInt("page", page); err != nil {
		return err
	}
	if err := utils.ValidatePositiveInt("limit", limit); err != nil {
		return err
	}
	return nil
}
