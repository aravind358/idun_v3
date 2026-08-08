package output

import (
	"context"
	
	"idun/intelligence/executive/v3"
	"idun/world"
)

// OutputManager orchestrates the asynchronous output egress pipeline.
type OutputManager interface {
	// Dispatch begins the asynchronous output process for a given ExecutionResult.
	// It is guaranteed not to block the caller.
	Dispatch(ctx context.Context, interaction *world.Interaction, execResult *v3.ExecutionResult) error
}
