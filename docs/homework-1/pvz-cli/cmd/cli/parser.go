package cli

import (
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

func (p *ArgsParser) AcceptOrderParams() handlers.AcceptOrderParams {
	m := p.asMap()
	return handlers.AcceptOrderParams{
		OrderID:   m["--order-id"],
		UserID:    m["--user-id"],
		ExpiresAt: m["--expires"],
	}
}

func (p *ArgsParser) ReturnOrderParams() handlers.ReturnOrderParams {
	m := p.asMap()
	return handlers.ReturnOrderParams{
		OrderID: m["--order-id"],
	}
}

func (p *ArgsParser) ProcessOrdersParams() handlers.ProcessOrdersParams {
	m := p.asMap()
	return handlers.ProcessOrdersParams{
		UserID:   m["--user-id"],
		Action:   m["--action"],
		OrderIDs: m["--order-ids"],
	}
}

func (p *ArgsParser) ListOrdersParams() handlers.ListOrdersParams {
	m := p.asMap()

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
	}
}

func (p *ArgsParser) ListReturnsParams() handlers.ListReturnsParams {
	m := p.asMap()
	return handlers.ListReturnsParams{
		Page:  parseOptionalInt(m["--page"]),
		Limit: parseOptionalInt(m["--limit"]),
	}
}

func (p *ArgsParser) ImportOrdersParams() handlers.ImportOrdersParams {
	m := p.asMap()
	return handlers.ImportOrdersParams{
		File: m["--file"],
	}
}

func (p *ArgsParser) ScrollOrdersParams() handlers.ScrollOrdersParams {
	m := p.asMap()
	return handlers.ScrollOrdersParams{
		UserID: m["--user-id"],
		Limit:  parseOptionalInt(m["--limit"]),
	}
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
