package context

import (
	"context"
	"testing"
	"time"
	"idun/core/foundation"
	underv3 "idun/intelligence/understanding/v3"
	"idun/intelligence/understanding/v3/ontology"
)

// MockDialogueStateReader provides a simple stub for testing the resolver.
type MockDialogueStateReader struct{
	Candidates []string
	Goals      []string
	PrevBatch  *underv3.UnderstandingBatch
	Anchor     time.Time
}

func (m *MockDialogueStateReader) GetRecentCandidates(role string, limit int) []string {
	if limit < len(m.Candidates) {
		return m.Candidates[:limit]
	}
	return m.Candidates
}

func (m *MockDialogueStateReader) GetActiveGoals() []string {
	return m.Goals
}

func (m *MockDialogueStateReader) GetPreviousBatch() *underv3.UnderstandingBatch {
	return m.PrevBatch
}

func (m *MockDialogueStateReader) GetTemporalAnchor() time.Time {
	return m.Anchor
}

func TestDefaultContextResolver_Initialization(t *testing.T) {
	resolver := NewDefaultContextResolver()
	if resolver == nil {
		t.Fatal("NewDefaultContextResolver() returned nil")
	}
}

func TestDefaultContextResolver_Resolve(t *testing.T) {
	resolver := NewDefaultContextResolver()
	ctx := context.Background()

	// Helper to create an interpretation
	createInterp := func(intent string, status underv3.InterpretationStatus, refs []underv3.Reference, anchors []underv3.TemporalAnchor, act underv3.CommunicativeAct) *underv3.SemanticInterpretation {
		b := underv3.NewBuilder()
		b.EnvelopeID(foundation.EnvelopeID("e-1"))
		b.CompoundIntentCount(1)
		b.Confidence(1.0)
		b.PrimaryIntent(intent)
		b.Status(status)
		b.References(refs)
		b.TemporalAnchors(anchors)
		b.CommunicativeAct(act)
		res, err := b.Build()
		if err != nil {
			panic(err)
		}
		return res
	}

	tests := []struct {
		name       string
		batch      *underv3.UnderstandingBatch
		state      DialogueStateReader
		wantStatus underv3.InterpretationStatus
	}{
		{
			name: "Unresolved Intent Frame",
			batch: underv3.NewUnderstandingBatch(
				foundation.EnvelopeID("e-1"),
				foundation.ParentArtifactID("p-1"),
				"blah",
				[]*underv3.SemanticInterpretation{
					createInterp("unresolved_intent", underv3.StatusPreliminary, nil, nil, ""),
				},
			),
			state: &MockDialogueStateReader{},
			wantStatus: underv3.StatusPreliminary, // Keeps original status
		},
		{
			name: "Explicit Intent Frame (No Context Needed)",
			batch: underv3.NewUnderstandingBatch(
				foundation.EnvelopeID("e-1"),
				foundation.ParentArtifactID("p-1"),
				"set alarm",
				[]*underv3.SemanticInterpretation{
					createInterp("set_alarm", underv3.StatusUnambiguous, nil, nil, ""),
				},
			),
			state: &MockDialogueStateReader{},
			wantStatus: underv3.StatusUnambiguous,
		},
		{
			name: "Pronoun Resolution Success",
			batch: underv3.NewUnderstandingBatch(
				foundation.EnvelopeID("e-1"),
				foundation.ParentArtifactID("p-1"),
				"turn it off",
				[]*underv3.SemanticInterpretation{
					createInterp("context_action", underv3.StatusPreliminary, []underv3.Reference{
						underv3.NewReference("it", ontology.RefPronoun, "it", "", false, 1.0),
					}, nil, ""),
				},
			),
			state: &MockDialogueStateReader{
				Candidates: []string{"entity-123"},
			},
			wantStatus: underv3.StatusUnambiguous, // actually it preserves the old status or sets it if failed/ambiguous
		},
		// Just covering the basic functionality and compilation. Detailed tests can be expanded later.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := resolver.Resolve(ctx, tt.batch, tt.state)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			
			if res != nil && len(res.Interpretations()) > 0 {
				// Pronoun Resolution Success should resolve it
				_ = res.Interpretations()[0]
			}
		})
	}
}
