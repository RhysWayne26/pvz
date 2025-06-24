package mappers

var _ GRPCFacadeMapper = (*DefaultGRPCFacadeMapper)(nil)

// DefaultGRPCFacadeMapper is a default implementation of GRPCFacadeMapper
type DefaultGRPCFacadeMapper struct{}

// NewDefaultGRPCFacadeMapper returns a new instance of DefaultGRPCFacadeMapper
func NewDefaultGRPCFacadeMapper() *DefaultGRPCFacadeMapper {
	return &DefaultGRPCFacadeMapper{}
}
