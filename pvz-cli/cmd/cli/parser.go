package cli

import (
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/dto"
	"strconv"
)

// ArgsParser parses command-line arguments into structured parameters for different commands
type ArgsParser struct {
	args []string
}

// NewArgsParser creates a new command-line arguments parser
func NewArgsParser(args []string) *ArgsParser {
	return &ArgsParser{args: args}
}

func (p *ArgsParser) asMap() map[string]string {
	m := make(map[string]string)
	i := 0
	for i < len(p.args) {
		key := p.args[i]
		if i+1 < len(p.args) && p.args[i+1][0] != '-' {
			m[key] = p.args[i+1]
			i += 2
		} else {
			m[key] = ""
			i++
		}
	}
	return m
}

// AcceptOrderParams parses and validates parameters for accept-order command
func (p *ArgsParser) AcceptOrderParams() (dto.AcceptOrderParams, error) {
	m := p.asMap()

	if m["--order-id"] == "" {
		return dto.AcceptOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "order-id is required")
	}
	if m["--user-id"] == "" {
		return dto.AcceptOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "user-id is required")
	}
	if m["--expires"] == "" {
		return dto.AcceptOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "expires is required")
	}
	if m["--weight"] == "" {
		return dto.AcceptOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "weight is required")
	}
	if m["--price"] == "" {
		return dto.AcceptOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "price is required")
	}

	return dto.AcceptOrderParams{
		OrderID:   m["--order-id"],
		UserID:    m["--user-id"],
		ExpiresAt: m["--expires"],
		Weight:    m["--weight"],
		Price:     m["--price"],
		Package:   m["--package"],
	}, nil
}

// ReturnOrderParams parses and validates parameters for return-order command
func (p *ArgsParser) ReturnOrderParams() (dto.ReturnOrderParams, error) {
	m := p.asMap()

	if m["--order-id"] == "" {
		return dto.ReturnOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "order-id is required")
	}

	return dto.ReturnOrderParams{
		OrderID: m["--order-id"],
	}, nil
}

// ProcessOrdersParams parses and validates parameters for process-orders command
func (p *ArgsParser) ProcessOrdersParams() (dto.ProcessOrdersParams, error) {
	m := p.asMap()

	if m["--user-id"] == "" {
		return dto.ProcessOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "user-id is required")
	}

	if m["--action"] == "" {
		return dto.ProcessOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "action is required")
	}

	if m["--order-ids"] == "" {
		return dto.ProcessOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "order-ids is required")
	}

	return dto.ProcessOrdersParams{
		UserID:   m["--user-id"],
		Action:   m["--action"],
		OrderIDs: m["--order-ids"],
	}, nil
}

// ListOrdersParams parses and validates parameters for list-orders command
func (p *ArgsParser) ListOrdersParams() (dto.ListOrdersParams, error) {
	m := p.asMap()
	allowed := map[string]struct{}{
		"--user-id": {}, "--in-pvz": {}, "--last": {},
		"--page": {}, "--limit": {}, "--last-id": {},
	}
	for key := range m {
		if _, ok := allowed[key]; !ok {
			return dto.ListOrdersParams{},
				apperrors.Newf(apperrors.ValidationFailed, "unknown flag %q", key)
		}
	}
	if m["--user-id"] == "" {
		return dto.ListOrdersParams{},
			apperrors.Newf(apperrors.ValidationFailed, "user-id is required")
	}

	inPvz, err := parseOptionalBool(m, "--in-pvz")
	if err != nil {
		return dto.ListOrdersParams{}, err
	}

	last, err := parseOptionalInt(m, "--last")
	if err != nil {
		return dto.ListOrdersParams{}, err
	}
	page, err := parseOptionalInt(m, "--page")
	if err != nil {
		return dto.ListOrdersParams{}, err
	}
	limit, err := parseOptionalInt(m, "--limit")
	if err != nil {
		return dto.ListOrdersParams{}, err
	}

	return dto.ListOrdersParams{
		UserID: m["--user-id"],
		InPvz:  inPvz,
		Last:   last,
		Page:   page,
		Limit:  limit,
		LastID: m["--last-id"],
	}, nil
}

// ListReturnsParams parses and validates parameters for list-returns command
func (p *ArgsParser) ListReturnsParams() (dto.ListReturnsParams, error) {
	m := p.asMap()

	page, err := parseOptionalInt(m, "--page")
	if err != nil {
		return dto.ListReturnsParams{}, err
	}
	limit, err := parseOptionalInt(m, "--limit")
	if err != nil {
		return dto.ListReturnsParams{}, err
	}

	return dto.ListReturnsParams{
		Page:  page,
		Limit: limit,
	}, nil
}

// ImportOrdersParams parses and validates parameters for import-orders command
func (p *ArgsParser) ImportOrdersParams() (dto.ImportOrdersParams, error) {
	m := p.asMap()

	if m["--file"] == "" {
		return dto.ImportOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "file is required")
	}

	return dto.ImportOrdersParams{
		File: m["--file"],
	}, nil
}

// ScrollOrdersParams parses and validates parameters for scroll-orders command
func (p *ArgsParser) ScrollOrdersParams() (dto.ScrollOrdersParams, error) {
	m := p.asMap()

	if m["--user-id"] == "" {
		return dto.ScrollOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "user-id is required")
	}

	limit, err := parseOptionalInt(m, "--limit")
	if err != nil {
		return dto.ScrollOrdersParams{}, err
	}

	return dto.ScrollOrdersParams{
		UserID: m["--user-id"],
		Limit:  limit,
	}, nil
}

func parseOptionalInt(m map[string]string, key string) (*int, error) {
	s, ok := m[key]
	if !ok || s == "" {
		return nil, nil
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return nil, apperrors.Newf(apperrors.ValidationFailed, "%s must be an integer", key)
	}
	return &val, nil
}

func parseOptionalBool(m map[string]string, key string) (*bool, error) {
	if _, ok := m[key]; !ok {
		return nil, nil
	}
	b := true
	if val := m[key]; val != "" {
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return nil, apperrors.Newf(apperrors.ValidationFailed, "%s must be boolean", key)
		}
		b = parsed
	}
	return &b, nil
}
