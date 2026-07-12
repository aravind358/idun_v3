package understanding_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"idun/intelligence/understanding"
)

func TestSlotValidation(t *testing.T) {
	valid := understanding.Slot{
		Name:        "destination",
		Value:       "Portland, OR",
		GroundingID: "geo-portland-or",
		Confidence:  0.95,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid slot, got: %v", err)
	}

	invalidName := valid
	invalidName.Name = ""
	if !errors.Is(invalidName.Validate(), understanding.ErrInvalidSlotName) {
		t.Fatalf("expected ErrInvalidSlotName for empty name")
	}

	invalidConf := valid
	invalidConf.Confidence = 1.5
	if !errors.Is(invalidConf.Validate(), understanding.ErrInvalidConfidence) {
		t.Fatalf("expected ErrInvalidConfidence for out-of-range confidence")
	}
}

func TestHypothesisValidationAndClone(t *testing.T) {
	hyp := understanding.Hypothesis{
		Intent:               "book_flight",
		CalibratedConfidence: 0.85,
		SourceLayer:          understanding.LayerNeuralClassifier,
		Slots: []understanding.Slot{
			{Name: "dest", Value: "PDX", Confidence: 0.90},
		},
	}
	if err := hyp.Validate(); err != nil {
		t.Fatalf("expected valid hypothesis, got: %v", err)
	}

	cloned := hyp.Clone()
	cloned.Slots[0].Value = "MUTATED"
	if hyp.Slots[0].Value != "PDX" {
		t.Fatalf("expected Clone to perform deep copy of Slots")
	}

	invalidIntent := hyp
	invalidIntent.Intent = ""
	if !errors.Is(invalidIntent.Validate(), understanding.ErrMissingIntent) {
		t.Fatalf("expected ErrMissingIntent")
	}
}

func TestSemanticFrameValidationAndBeamOverflow(t *testing.T) {
	builder := understanding.NewSemanticFrameBuilder("env-100")
	builder.WithPrimaryHypothesis("reschedule", 0.90, understanding.LayerReflexiveGrammar)

	// Add 2 ambiguous runner-up hypotheses -> total 3 (MaxBeamWidth)
	builder.AddAmbiguousHypothesis("reschedule_meeting", 0.88, understanding.LayerNeuralClassifier, 0.02)
	builder.AddAmbiguousHypothesis("cancel_meeting", 0.85, understanding.LayerNeuralClassifier, 0.05)

	frame, err := builder.Build()
	if err != nil {
		t.Fatalf("expected build success for beam width 3, got: %v", err)
	}
	if frame.Status != understanding.StatusAmbiguousBeam {
		t.Fatalf("expected status AMBIGUOUS_BEAM, got %s", frame.Status)
	}

	// Attempt adding a 3rd runner-up -> total 4 -> must return ErrBeamOverflow
	builder.AddAmbiguousHypothesis("create_meeting", 0.80, understanding.LayerNeuralClassifier, 0.10)
	_, err = builder.Build()
	if !errors.Is(err, understanding.ErrBeamOverflow) {
		t.Fatalf("expected ErrBeamOverflow when exceeding MaxBeamWidth, got: %v", err)
	}
}

func TestSemanticFrameJSONSerialization(t *testing.T) {
	builder := understanding.NewSemanticFrameBuilder("env-404").
		WithStatus(understanding.StatusUnambiguous).
		WithPrimaryHypothesis("query_weather", 0.98, understanding.LayerReflexiveGrammar,
			understanding.Slot{Name: "city", Value: "Seattle", GroundingID: "geo-seattle", Confidence: 0.99},
		).
		WithProcessedDuration(4.2)

	frame, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"AmbiguitySet":[]`) {
		t.Fatalf("expected empty non-nil AmbiguitySet array in JSON output, got: %s", jsonStr)
	}

	var decoded understanding.SemanticFrame
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("decoded frame failed validation: %v", err)
	}
	if decoded.EnvelopeID != "env-404" || decoded.PrimaryHypothesis.Intent != "query_weather" {
		t.Fatalf("decoded frame mismatch")
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := understanding.WithConfigOptions()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected default config valid, got: %v", err)
	}

	invalidCfg := cfg
	invalidCfg.MaxBeamWidth = 0
	if invalidCfg.Validate() == nil {
		t.Fatalf("expected error for invalid MaxBeamWidth")
	}
}
