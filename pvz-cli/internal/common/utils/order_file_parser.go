package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/common/apperrors"
	"strings"
)

// ParseOrdersFromFile reads and parses order data from JSON file into order requests
func ParseOrdersFromFile(filePath string) ([]params.AcceptOrderParams, error) {
	if err := validateFilePath(filePath); err != nil {
		return nil, err
	}

	// #nosec G304 -- filePath is validated by validateFilePath() right above
	f, err := os.Open(filePath)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "cannot open file %q: %v", filePath, err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()

	var orders []params.AcceptOrderParams
	if err := json.NewDecoder(f).Decode(&orders); err != nil {
		return nil, apperrors.Newf(apperrors.ValidationFailed, "invalid JSON: %v", err)
	}

	return orders, nil
}

func validateFilePath(filePath string) error {
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed")
	}
	return nil
}
