package reasoning

import (
	"context"

	"idun/core/memory"
	"idun/intelligence/communication"
)

// MemoryProvider defines the minimal interface required from idun/core/memory
// for Stage S0 Context & Strategy Assembly.
type MemoryProvider interface {
	ListRecordsByType(recordType string) ([]memory.Record, error)
}

// ContextAssembler defines the interface for Stage S0 Context Assembly.
// It retrieves a bounded, activation-ranked slice of Memory records as working context.
type ContextAssembler interface {
	AssembleContext(ctx context.Context, perceptionEnv communication.Envelope, maxRecords int) ([]memory.Record, error)
}

// DefaultContextAssembler implements ContextAssembler backed by a MemoryProvider.
type DefaultContextAssembler struct {
	mem MemoryProvider
}

// NewDefaultContextAssembler returns a new DefaultContextAssembler.
// mem may be nil when running purely local/in-memory reasoning.
func NewDefaultContextAssembler(mem MemoryProvider) *DefaultContextAssembler {
	return &DefaultContextAssembler{mem: mem}
}

// AssembleContext retrieves up to maxRecords relevant records from Memory.
// It queries canonical knowledge types ("belief", "fact") and bounds the returned slice
// so full historical memory is never loaded into working memory.
func (a *DefaultContextAssembler) AssembleContext(ctx context.Context, perceptionEnv communication.Envelope, maxRecords int) ([]memory.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if a.mem == nil || maxRecords <= 0 {
		return []memory.Record{}, nil
	}

	records := make([]memory.Record, 0, maxRecords)

	for _, recType := range []string{"belief", "fact"} {
		found, err := a.mem.ListRecordsByType(recType)
		if err != nil {
			continue
		}
		for _, rec := range found {
			if len(records) >= maxRecords {
				break
			}
			records = append(records, rec)
		}
		if len(records) >= maxRecords {
			break
		}
	}

	return records, nil
}
