package understanding

import (
	"context"
	"testing"

	"idun/intelligence/communication"
)

func TestContextGrammarRules(t *testing.T) {
	svc := NewService(Config{}, nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		input      string
		wantIntent string
		wantSlot   string
		slotName   string
	}{
		{
			name:       "Context Action - Delete it",
			input:      "delete it",
			wantIntent: "context_action",
			wantSlot:   "it",
			slotName:   "Anaphora", // from ConceptAnaphora Category="Slot"
		},
		{
			name:       "Context Action - Read that",
			input:      "read that",
			wantIntent: "context_action",
			wantSlot:   "that",
			slotName:   "Anaphora",
		},
		{
			name:       "Context Ellipsis - what about battery",
			input:      "what about battery",
			wantIntent: "context_ellipsis",
			wantSlot:   "battery",
			slotName:   "concept",
		},
		{
			name:       "Context Ellipsis - and tomorrow",
			input:      "and tomorrow",
			wantIntent: "context_ellipsis",
			wantSlot:   "tomorrow",
			slotName:   "concept",
		},
		{
			name:       "Confirmation - yes",
			input:      "yes",
			wantIntent: "confirmation",
		},
		{
			name:       "Confirmation - do it",
			input:      "do it",
			wantIntent: "confirmation",
		},
		{
			name:       "Negation - no",
			input:      "no",
			wantIntent: "negation",
		},
		{
			name:       "Negation - cancel",
			input:      "cancel",
			wantIntent: "negation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := svc.InterpretEnvelope(ctx, communication.Envelope{
				ID:         "test-" + tt.name,
				Topic:      communication.TopicPerception,
				PayloadRef: tt.input,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if frame.PrimaryHypothesis.Intent != tt.wantIntent {
				t.Errorf("got intent %q, want %q", frame.PrimaryHypothesis.Intent, tt.wantIntent)
			}

			if tt.slotName != "" {
				found := false
				for _, slot := range frame.PrimaryHypothesis.Slots {
					if slot.Name == tt.slotName {
						found = true
						if slot.Value != tt.wantSlot {
							t.Errorf("got slot value %q, want %q", slot.Value, tt.wantSlot)
						}
					}
				}
				if !found {
					t.Errorf("expected slot %q not found. Slots: %v", tt.slotName, frame.PrimaryHypothesis.Slots)
				}
			}
		})
	}
}
