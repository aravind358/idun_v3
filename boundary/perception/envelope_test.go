package perception

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPerceptionEnvelope_Builder(t *testing.T) {
	now := time.Now()
	env, err := NewBuilder().
		EnvelopeID("env-123").
		ArtifactID("art-456").
		Version("1.0").
		Timestamp(now).
		InputSource("voice").
		InputType("audio").
		RawInput("hello idun").
		Metadata("key", "value").
		Build()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(env.EnvelopeID()) != "env-123" {
		t.Errorf("expected env-123, got %s", env.EnvelopeID())
	}
	if env.RawInput() != "hello idun" {
		t.Errorf("expected 'hello idun', got %s", env.RawInput())
	}
	if env.Metadata()["key"] != "value" {
		t.Errorf("expected 'value', got %v", env.Metadata()["key"])
	}

	// Verify immutability of metadata map via getter
	meta := env.Metadata()
	meta["key"] = "hacked"
	if env.Metadata()["key"] == "hacked" {
		t.Errorf("expected internal metadata to be protected from mutation")
	}
}

func TestPerceptionEnvelope_Validation(t *testing.T) {
	_, err := NewBuilder().Build()
	if err == nil {
		t.Errorf("expected error building empty envelope")
	}

	_, err = NewBuilder().
		EnvelopeID("env-123").
		ArtifactID("art-456").
		Version("1.0").
		Build() // Missing raw input
	if err == nil {
		t.Errorf("expected error for missing raw input")
	}
}

func TestPerceptionEnvelope_Serialization(t *testing.T) {
	now := time.Now().Round(time.Millisecond) // Round for consistent JSON serialization comparison
	env, _ := NewBuilder().
		EnvelopeID("env-123").
		ArtifactID("art-456").
		Version("1.0").
		Timestamp(now).
		InputSource("voice").
		InputType("audio").
		RawInput("hello idun").
		Build()

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var env2 PerceptionEnvelope
	if err := json.Unmarshal(data, &env2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if env2.EnvelopeID() != env.EnvelopeID() {
		t.Errorf("mismatch EnvelopeID: got %s, want %s", env2.EnvelopeID(), env.EnvelopeID())
	}
	if time.Time(env2.Timestamp()).UnixMilli() != time.Time(env.Timestamp()).UnixMilli() {
		t.Errorf("mismatch Timestamp: got %v, want %v", env2.Timestamp(), env.Timestamp())
	}
}
