package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// ParseOrdersFromFile reads and parses order data from JSON file into order requests
func ParseOrdersFromFile(filePath string) ([]requests.AcceptOrderRequest, error) {
	rawData, err := readRawData(filePath)
	if err != nil {
		return nil, err
	}

	return parseRawData(rawData), nil
}

func validateFilePath(filePath string) error {
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed")
	}
	return nil
}

func readRawData(filePath string) ([]map[string]string, error) {
	if err := validateFilePath(filePath); err != nil {
		return nil, err
	}

	// #nosec G304 -- filePath is validated by validateFilePath() right above
	f, err := os.Open(filePath)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "cannot open file %q: %v", filePath, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("WARNING: failed to close file: %v\n", err)
		}
	}()

	var raw []map[string]string
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, apperrors.Newf(apperrors.ValidationFailed, "invalid JSON: %v", err)
	}

	return raw, nil
}

func parseRawData(raw []map[string]string) []requests.AcceptOrderRequest {
	orderRequests := make([]requests.AcceptOrderRequest, 0)

	for _, item := range raw {
		req, err := parseOrderItem(item)
		if err != nil {
			continue
		}
		orderRequests = append(orderRequests, req)
	}

	return orderRequests
}

func parseOrderItem(item map[string]string) (requests.AcceptOrderRequest, error) {
	orderID := strings.TrimSpace(item["order_id"])
	userID := strings.TrimSpace(item["user_id"])

	expiresAt, err := time.Parse(constants.TimeLayout, item["expires_at"])
	if err != nil {
		fmt.Printf("SKIPPED: %s — invalid expires_at\n", orderID)
		return requests.AcceptOrderRequest{}, err
	}

	weight, err := parseFloat(item["weight"])
	if err != nil {
		fmt.Printf("ERROR importing %s: invalid weight\n", orderID)
		return requests.AcceptOrderRequest{}, err
	}

	price, err := parseFloat(item["price"])
	if err != nil {
		fmt.Printf("ERROR importing %s: invalid price\n", orderID)
		return requests.AcceptOrderRequest{}, err
	}

	pkg := parsePackageType(item["package"])

	return requests.AcceptOrderRequest{
		OrderID:   orderID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Weight:    weight,
		Price:     price,
		Package:   pkg,
	}, nil
}

func parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}

func parsePackageType(packageStr string) models.PackageType {
	trimmed := strings.TrimSpace(packageStr)
	if trimmed == "" {
		return models.PackageNone
	}
	return models.PackageType(trimmed)
}
