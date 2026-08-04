package v3

import (
	"context"
	"idun/core/memory"
)

// LegacyMemoryAdapter wraps the core memory service to implement the V3 MemoryProvider abstraction.
type LegacyMemoryAdapter struct {
	mem memory.Memory
}

// NewLegacyMemoryAdapter creates a new memory adapter.
func NewLegacyMemoryAdapter(mem memory.Memory) MemoryProvider {
	return &LegacyMemoryAdapter{mem: mem}
}

func (a *LegacyMemoryAdapter) RetrieveEntity(ctx context.Context, surfaceName string) (memoryID string, confidence float64, err error) {
	// The V1 memory system does not natively support semantic entity retrieval by surface name.
	// For integration purposes, we return a mock empty response. A full V3 architecture
	// would require an embeddings-based semantic index here.
	return "", 0.0, nil
}

func (a *LegacyMemoryAdapter) ResolveReference(ctx context.Context, pronoun string) (targetSurface string, memoryID string, confidence float64, err error) {
	return "", "", 0.0, nil
}

func (a *LegacyMemoryAdapter) RetrieveContext(ctx context.Context, intent string, topics []string) ([]ContextEvidence, error) {
	// Retrieve facts/beliefs directly to satisfy ContextEvidence requirements.
	var evidence []ContextEvidence
	if a.mem == nil {
		return evidence, nil
	}

	for _, recType := range []string{"belief", "fact"} {
		found, err := a.mem.ListRecordsByType(recType)
		if err != nil {
			continue
		}
		for i, rec := range found {
			if i >= 3 {
				break
			}
			evidence = append(evidence, NewContextEvidence(rec.ID, string(rec.Payload), 0.8))
		}
	}
	return evidence, nil
}

func (a *LegacyMemoryAdapter) EvaluateCondition(ctx context.Context, condition string) (bool, error) {
	return true, nil // Optimistically allow conditions for integration.
}

func (a *LegacyMemoryAdapter) EvaluateFact(ctx context.Context, premise string) (bool, error) {
	return true, nil // Optimistically allow facts for integration.
}
