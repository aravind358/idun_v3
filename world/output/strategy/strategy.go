package strategy

import (
	"idun/world/output"
)

// DefaultStrategy implements the Strategy interface by looking up 
// Realization Descriptors in a simple registry.
type DefaultStrategy struct {
	registry map[output.ResponseType]output.Descriptor
	fallback output.Descriptor
}

// NewDefaultStrategy constructs a DefaultStrategy with a mandatory fallback.
func NewDefaultStrategy(fallback output.Descriptor) *DefaultStrategy {
	return &DefaultStrategy{
		registry: make(map[output.ResponseType]output.Descriptor),
		fallback: fallback,
	}
}

// Register adds a Descriptor mapping for a specific ResponseType.
func (s *DefaultStrategy) Register(rt output.ResponseType, desc output.Descriptor) {
	s.registry[rt] = desc
}

// Select retrieves the configured Descriptor for the given ResponseType.
func (s *DefaultStrategy) Select(rt output.ResponseType) output.Descriptor {
	if desc, ok := s.registry[rt]; ok {
		return desc
	}
	return s.fallback
}
