package formatter

import (
	"context"
	
	"idun/world/output"
)

// Formatter defines a generic abstraction for modality-specific document formatting.
type Formatter interface {
	Format(ctx context.Context, doc output.OutputDocument) error
}
