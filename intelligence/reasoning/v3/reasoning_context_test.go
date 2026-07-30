package v3

import (
	"encoding/json"
	"idun/core/foundation"
	understanding "idun/intelligence/understanding/v3"
	"testing"
	"time"
)

func TestReasoningContext_BuilderAndValidation(t *testing.T) {
	artID, _ := foundation.NewUUID()
	parentID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	tests := []struct {
		name    string
		builder *Builder
		wantErr bool
	}{
		{
			name: "valid context",
			builder: NewBuilder().
				ArtifactID(foundation.ArtifactID(artID)).
				ParentArtifactID(foundation.ParentArtifactID(parentID)).
				EnvelopeID(foundation.EnvelopeID(envID)).
				Timestamp(foundation.Timestamp(time.Now())).
				ResolvedIntent("resolved_intent"),
			wantErr: false,
		},
		{
			name: "missing artifact ID",
			builder: NewBuilder().
				ParentArtifactID(foundation.ParentArtifactID(parentID)).
				EnvelopeID(foundation.EnvelopeID(envID)).
				Timestamp(foundation.Timestamp(time.Now())).
				ResolvedIntent("resolved_intent"),
			wantErr: true,
		},
		{
			name: "missing parent ID",
			builder: NewBuilder().
				ArtifactID(foundation.ArtifactID(artID)).
				EnvelopeID(foundation.EnvelopeID(envID)).
				Timestamp(foundation.Timestamp(time.Now())).
				ResolvedIntent("resolved_intent"),
			wantErr: true,
		},
		{
			name: "missing envelope ID",
			builder: NewBuilder().
				ArtifactID(foundation.ArtifactID(artID)).
				ParentArtifactID(foundation.ParentArtifactID(parentID)).
				Timestamp(foundation.Timestamp(time.Now())).
				ResolvedIntent("resolved_intent"),
			wantErr: true,
		},
		{
			name: "missing resolved intent",
			builder: NewBuilder().
				ArtifactID(foundation.ArtifactID(artID)).
				ParentArtifactID(foundation.ParentArtifactID(parentID)).
				EnvelopeID(foundation.EnvelopeID(envID)).
				Timestamp(foundation.Timestamp(time.Now())),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReasoningContext_Serialization(t *testing.T) {
	artID, _ := foundation.NewUUID()
	parentID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	origSlot := understanding.NewSlot("time", "tomorrow", "", 0.9)
	enriched := NewEnrichedSlot(origSlot, "2026-07-26T00:00:00Z")

	ctx, err := NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		ParentArtifactID(foundation.ParentArtifactID(parentID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		ResolvedIntent("book_meeting").
		EnrichedSlots([]EnrichedSlot{enriched}).
		GroundedEntities([]GroundedEntity{NewGroundedEntity("John", "kb-john-doe", 0.95)}).
		ResolvedReferences([]ResolvedReference{NewResolvedReference("him", "John", "kb-john-doe", 0.9)}).
		RetrievedContexts([]ContextEvidence{NewContextEvidence("semantic", "John is a VIP", 0.88)}).
		ConditionEvaluated(true, false).
		TruthEvaluated(true, true).
		Build()

	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	data, err := json.Marshal(ctx)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var ctx2 ReasoningContext
	if err := json.Unmarshal(data, &ctx2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if ctx2.ArtifactID() != ctx.ArtifactID() {
		t.Errorf("mismatch artifact ID")
	}
	if ctx2.EnrichedSlots()[0].EnrichedValue() != "2026-07-26T00:00:00Z" {
		t.Errorf("mismatch enriched slot")
	}
	if ctx2.ConditionMet() {
		t.Errorf("expected condition met false")
	}
	if !ctx2.IsFactuallyTrue() {
		t.Errorf("expected factually true")
	}
}
