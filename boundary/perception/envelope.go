package perception

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"idun/core/foundation"
)

var (
	ErrMissingEnvelopeID = errors.New("perception: envelope ID is required")
	ErrMissingArtifactID = errors.New("perception: artifact ID is required")
	ErrMissingVersion    = errors.New("perception: version is required")
	ErrMissingRawInput   = errors.New("perception: raw input is required")
)

// PerceptionEnvelope represents the immutable input boundary from the World layer.
type PerceptionEnvelope struct {
	envelopeID       foundation.EnvelopeID
	artifactID       foundation.ArtifactID
	parentArtifactID foundation.ParentArtifactID
	version          foundation.Version
	timestamp        foundation.Timestamp
	inputSource      string
	inputType        string
	rawInput         string
	metadata         map[string]any
}

// EnvelopeID returns the golden thread correlation ID.
func (p *PerceptionEnvelope) EnvelopeID() foundation.EnvelopeID {
	return p.envelopeID
}

// ArtifactID returns the unique ID of this envelope artifact.
func (p *PerceptionEnvelope) ArtifactID() foundation.ArtifactID {
	return p.artifactID
}

// ParentArtifactID returns the ID of the upstream artifact (typically empty for the PerceptionEnvelope).
func (p *PerceptionEnvelope) ParentArtifactID() foundation.ParentArtifactID {
	return p.parentArtifactID
}

// Version returns the schema version.
func (p *PerceptionEnvelope) Version() foundation.Version {
	return p.version
}

// Timestamp returns the time this perception occurred.
func (p *PerceptionEnvelope) Timestamp() foundation.Timestamp {
	return p.timestamp
}

// InputSource returns the origin of the input (e.g., "voice", "text_ui").
func (p *PerceptionEnvelope) InputSource() string {
	return p.inputSource
}

// InputType returns the type of the input (e.g., "audio", "text").
func (p *PerceptionEnvelope) InputType() string {
	return p.inputType
}

// RawInput returns the verbatim string representation of the perception.
func (p *PerceptionEnvelope) RawInput() string {
	return p.rawInput
}

// Metadata returns any non-semantic telemetry or context provided by the World layer.
// Returns a copy to prevent mutation of the internal map.
func (p *PerceptionEnvelope) Metadata() map[string]any {
	if p.metadata == nil {
		return nil
	}
	cpy := make(map[string]any, len(p.metadata))
	for k, v := range p.metadata {
		cpy[k] = v
	}
	return cpy
}

// IsImmutable satisfies foundation.Immutable.
func (p *PerceptionEnvelope) IsImmutable() bool {
	return true
}

// Validate checks the structural invariants of the PerceptionEnvelope.
func (p *PerceptionEnvelope) Validate() error {
	if p.envelopeID == "" {
		return ErrMissingEnvelopeID
	}
	if p.artifactID == "" {
		return ErrMissingArtifactID
	}
	if p.version == "" {
		return ErrMissingVersion
	}
	if p.rawInput == "" {
		return ErrMissingRawInput
	}
	return nil
}

// Builder for PerceptionEnvelope
type Builder struct {
	env *PerceptionEnvelope
}

// NewBuilder initializes a builder for PerceptionEnvelope.
func NewBuilder() *Builder {
	return &Builder{
		env: &PerceptionEnvelope{
			metadata: make(map[string]any),
		},
	}
}

func (b *Builder) EnvelopeID(id string) *Builder {
	b.env.envelopeID = foundation.EnvelopeID(id)
	return b
}

func (b *Builder) ArtifactID(id string) *Builder {
	b.env.artifactID = foundation.ArtifactID(id)
	return b
}

func (b *Builder) ParentArtifactID(id string) *Builder {
	b.env.parentArtifactID = foundation.ParentArtifactID(id)
	return b
}

func (b *Builder) Version(v string) *Builder {
	b.env.version = foundation.Version(v)
	return b
}

func (b *Builder) Timestamp(t time.Time) *Builder {
	b.env.timestamp = foundation.Timestamp(t)
	return b
}

func (b *Builder) InputSource(s string) *Builder {
	b.env.inputSource = s
	return b
}

func (b *Builder) InputType(t string) *Builder {
	b.env.inputType = t
	return b
}

func (b *Builder) RawInput(r string) *Builder {
	b.env.rawInput = r
	return b
}

func (b *Builder) Metadata(k string, v any) *Builder {
	b.env.metadata[k] = v
	return b
}

// Build constructs and validates the PerceptionEnvelope.
func (b *Builder) Build() (*PerceptionEnvelope, error) {
	if err := b.env.Validate(); err != nil {
		return nil, fmt.Errorf("perception: build failed: %w", err)
	}
	return b.env, nil
}

// jsonEnvelope is used for serialization to map unexported fields.
type jsonEnvelope struct {
	EnvelopeID       string         `json:"envelope_id"`
	ArtifactID       string         `json:"artifact_id"`
	ParentArtifactID string         `json:"parent_artifact_id,omitempty"`
	Version          string         `json:"version"`
	Timestamp        time.Time      `json:"timestamp"`
	InputSource      string         `json:"input_source"`
	InputType        string         `json:"input_type"`
	RawInput         string         `json:"raw_input"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (p *PerceptionEnvelope) MarshalJSON() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("perception: cannot marshal invalid envelope: %w", err)
	}
	j := jsonEnvelope{
		EnvelopeID:       string(p.envelopeID),
		ArtifactID:       string(p.artifactID),
		ParentArtifactID: string(p.parentArtifactID),
		Version:          string(p.version),
		Timestamp:        time.Time(p.timestamp),
		InputSource:      p.inputSource,
		InputType:        p.inputType,
		RawInput:         p.rawInput,
		Metadata:         p.metadata,
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PerceptionEnvelope) UnmarshalJSON(data []byte) error {
	var j jsonEnvelope
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.envelopeID = foundation.EnvelopeID(j.EnvelopeID)
	p.artifactID = foundation.ArtifactID(j.ArtifactID)
	p.parentArtifactID = foundation.ParentArtifactID(j.ParentArtifactID)
	p.version = foundation.Version(j.Version)
	p.timestamp = foundation.Timestamp(j.Timestamp)
	p.inputSource = j.InputSource
	p.inputType = j.InputType
	p.rawInput = j.RawInput
	p.metadata = j.Metadata

	if err := p.Validate(); err != nil {
		return fmt.Errorf("perception: unmarshaled invalid envelope: %w", err)
	}
	return nil
}
