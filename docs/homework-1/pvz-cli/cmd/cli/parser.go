package cli

import (
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/cli/handlers"
	"strconv"
)

type ArgsParser struct {
	args []string
}

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

func (p *ArgsParser) AcceptOrderParams() (handlers.AcceptOrderParams, error) {
	m := p.asMap()

	if m["--order-id"] == "" {
		return handlers.AcceptOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "order-id is required")
	}

	if m["--user-id"] == "" {
		return handlers.AcceptOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "user-id is required")
	}

	if m["--expires"] == "" {
		return handlers.AcceptOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "expires is required")
	}

	return handlers.AcceptOrderParams{
		OrderID:   m["--order-id"],
		UserID:    m["--user-id"],
		ExpiresAt: m["--expires"],
	}, nil
}

func (p *ArgsParser) ReturnOrderParams() (handlers.ReturnOrderParams, error) {
	m := p.asMap()

	if m["--order-id"] == "" {
		return handlers.ReturnOrderParams{}, apperrors.Newf(apperrors.ValidationFailed, "order-id is required")
	}

	return handlers.ReturnOrderParams{
		OrderID: m["--order-id"],
	}, nil
}

func (p *ArgsParser) ProcessOrdersParams() (handlers.ProcessOrdersParams, error) {
	m := p.asMap()

	if m["--user-id"] == "" {
		return handlers.ProcessOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "user-id is required")
	}

	if m["--action"] == "" {
		return handlers.ProcessOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "action is required")
	}

	if m["--order-ids"] == "" {
		return handlers.ProcessOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "order-ids is required")
	}

	return handlers.ProcessOrdersParams{
		UserID:   m["--user-id"],
		Action:   m["--action"],
		OrderIDs: m["--order-ids"],
	}, nil
}

func (p *ArgsParser) ListOrdersParams() (handlers.ListOrdersParams, error) {
	m := p.asMap()

	if m["--user-id"] == "" {
		return handlers.ListOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "user-id is required")
	}

	_, hasInPvz := m["--in-pvz"]
	inPvz := hasInPvz

	return handlers.ListOrdersParams{
		UserID:   m["--user-id"],
		InPvz:    inPvz,
		UseInPvz: hasInPvz,
		Last:     parseOptionalInt(m["--last"]),
		Page:     parseOptionalInt(m["--page"]),
		Limit:    parseOptionalInt(m["--limit"]),
		LastID:   m["--last-id"],
	}, nil
}

func (p *ArgsParser) ListReturnsParams() (handlers.ListReturnsParams, error) {
	m := p.asMap()
	return handlers.ListReturnsParams{
		Page:  parseOptionalInt(m["--page"]),
		Limit: parseOptionalInt(m["--limit"]),
	}, nil
}

func (p *ArgsParser) ImportOrdersParams() (handlers.ImportOrdersParams, error) {
	m := p.asMap()

	if m["--file"] == "" {
		return handlers.ImportOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "file is required")
	}

	return handlers.ImportOrdersParams{
		File: m["--file"],
	}, nil
}

func (p *ArgsParser) ScrollOrdersParams() (handlers.ScrollOrdersParams, error) {
	m := p.asMap()

	if m["--user-id"] == "" {
		return handlers.ScrollOrdersParams{}, apperrors.Newf(apperrors.ValidationFailed, "user-id is required")
	}

	return handlers.ScrollOrdersParams{
		UserID: m["--user-id"],
		Limit:  parseOptionalInt(m["--limit"]),
	}, nil
}

func parseOptionalInt(s string) *int {
	if s == "" {
		return nil
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	return &val
}
