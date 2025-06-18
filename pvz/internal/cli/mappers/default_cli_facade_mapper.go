package mappers

var _ CLIFacadeMapper = (*DefaultCLIFacadeMapper)(nil)

// DefaultCLIFacadeMapper is a default implementation of CLIFacadeMapper
type DefaultCLIFacadeMapper struct{}

// NewDefaultFacadeMapper returns a new instance of DefaultCLIFacadeMapper
func NewDefaultFacadeMapper() *DefaultCLIFacadeMapper {
	return &DefaultCLIFacadeMapper{}
}
