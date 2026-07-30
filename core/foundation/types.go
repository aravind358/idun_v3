package foundation

import "time"

// ArtifactID is a unique identifier for a specific cognitive artifact instance.
type ArtifactID string

// ParentArtifactID is a unique identifier pointing to the upstream artifact that generated this one.
type ParentArtifactID string

// EnvelopeID is a unique identifier tying together all artifacts originating from
// a single World perception event (the golden thread).
type EnvelopeID string

// Version represents a canonical schema version string (e.g., "3.0").
type Version string

// Timestamp represents an absolute point in time, stored internally as time.Time.
type Timestamp time.Time

// String returns the Timestamp formatted as an ISO 8601 string.
func (t Timestamp) String() string {
	return time.Time(t).Format(time.RFC3339Nano)
}

// Validatable enforces self-validation on cognitive artifacts.
type Validatable interface {
	Validate() error
}

// Immutable marks an object as read-only.
// Any artifact published to the Global Workspace must implement this to 
// document its immutability guarantee.
type Immutable interface {
	IsImmutable() bool
}

// CognitiveArtifact defines the base lineage contract for all artifacts 
// flowing through the cognitive pipeline.
type CognitiveArtifact interface {
	Validatable
	Immutable
	ArtifactID() ArtifactID
	ParentArtifactID() ParentArtifactID
	EnvelopeID() EnvelopeID
	Timestamp() Timestamp
	Version() Version
}

