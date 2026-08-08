package output

import (
	"context"
	"idun/world"
)

// OutputPlugin defines the contract for capability-based modality plugins.
type OutputPlugin interface {
	Name() string
	SupportedModalities() []world.Modality
	Priority() int

	// Realize translates machine data into a format-agnostic OutputDocument.
	Realize(ctx context.Context, response CompositeResponse) (OutputDocument, error)
	
	// Format applies modality-specific formatting to the OutputDocument.
	Format(ctx context.Context, doc OutputDocument) error
	
	// Write performs physical I/O to deliver the formatted output.
	Write(ctx context.Context, output any) error
}
