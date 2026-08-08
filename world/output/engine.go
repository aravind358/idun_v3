package output

import (
	"context"
)

// OutputEngine converts a CompositeResponse into a structured OutputDocument.
type OutputEngine interface {
	Realize(ctx context.Context, response CompositeResponse) (OutputDocument, error)
}
