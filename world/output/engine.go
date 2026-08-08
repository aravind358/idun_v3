package output

import (
	"context"
)

// ResponseType indicates the logical classification of the output payload.
type ResponseType string

// OutputPayload is a neutral, world-owned container for capability data.
type OutputPayload struct {
	ResponseType ResponseType
	Data         any
}

// Descriptor provides the single source of configuration for realizing a ResponseType.
type Descriptor struct {
	ResponseType ResponseType
	Realizer     Realizer
	TemplateID   string
}

// Realizer implements the actual realization logic for a given response.
type Realizer interface {
	Realize(ctx context.Context, response CompositeResponse, desc Descriptor) (OutputDocument, error)
}

// Strategy selects the appropriate realization configuration.
type Strategy interface {
	Select(rt ResponseType) Descriptor
}

// OutputEngine orchestrates the realization pipeline using a Strategy.
type OutputEngine interface {
	Realize(ctx context.Context, response CompositeResponse) (OutputDocument, error)
}
