package reasoning

import (
	"context"
	"testing"

	"idun/core/memory"
	"idun/intelligence/communication"
)

type mockMemoryProvider struct {
	records map[string][]memory.Record
}

func (m *mockMemoryProvider) ListRecordsByType(recordType string) ([]memory.Record, error) {
	if recs, ok := m.records[recordType]; ok {
		return recs, nil
	}
	return []memory.Record{}, nil
}

func TestContextAssembler_NilMemory(t *testing.T) {
	assembler := NewDefaultContextAssembler(nil)
	records, err := assembler.AssembleContext(context.Background(), communication.Envelope{ID: "env-1"}, 10)
	if err != nil {
		t.Fatalf("expected nil memory assembler to succeed, got %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestContextAssembler_BoundedRetrieval(t *testing.T) {
	mem := &mockMemoryProvider{
		records: map[string][]memory.Record{
			"belief": {
				{ID: "bel-1", Type: "belief"},
				{ID: "bel-2", Type: "belief"},
			},
			"fact": {
				{ID: "fact-1", Type: "fact"},
				{ID: "fact-2", Type: "fact"},
			},
		},
	}

	assembler := NewDefaultContextAssembler(mem)
	records, err := assembler.AssembleContext(context.Background(), communication.Envelope{ID: "env-1"}, 3)
	if err != nil {
		t.Fatalf("expected bounded retrieval to succeed, got %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected exactly 3 records bounded by limit, got %d", len(records))
	}
}
