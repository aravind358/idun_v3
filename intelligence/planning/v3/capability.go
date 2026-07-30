package v3

import (
	"context"
)

// CapabilityDescriptor represents the abstract metadata of a capability,
// without exposing the executor, client, or any implementation details.
type CapabilityDescriptor struct {
	ID          CapabilityID
	Description string
	Params      []string // Required parameter names
}

func NewCapabilityDescriptor(id CapabilityID, desc string, params []string) CapabilityDescriptor {
	return CapabilityDescriptor{ID: id, Description: desc, Params: params}
}

// CapabilityRegistry provides the metadata of available capabilities to the planner.
type CapabilityRegistry interface {
	Discover(ctx context.Context, goal string) ([]CapabilityDescriptor, error)
}
