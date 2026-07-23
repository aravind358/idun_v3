// Package boundary defines the official interface between IDUN V3's Cognitive System
// and the external World layer (Text, Voice, GUI, API, and future modules).
//
// Architectural Rule: Not a Cognitive Object
// Boundary structures (such as CommunicationMessage) DO NOT belong to the internal
// cognitive hierarchy (Understanding, Reasoning, Planning, Decision, or Executive).
// They exist strictly at the edge to decouple internal cognitive schemas from external
// presentation and realization requirements.
package boundary

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrNilMessage        = errors.New("boundary: communication message is nil")
	ErrMissingResponseID = errors.New("boundary: communication message requires response_id")
)

// Slot represents an extracted semantic argument or referent binding at the boundary.
// It is mapped explicitly from internal cognitive slots to ensure downstream World modules
// do not depend on internal cognitive packages.
type Slot struct {
	Name        string  `json:"name"`
	Value       string  `json:"value"`
	GroundingID string  `json:"grounding_id"`
	Confidence  float64 `json:"confidence"`
}

// Entity represents canonical ontology or memory grounding details at the boundary.
type Entity struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	CanonicalName string            `json:"canonical_name"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

// CommunicationMessage represents everything the World layer needs in order to communicate
// with users or external systems. It exists only at the boundary between Cognition and World.
//
// Architectural Rule: Immutability
// After the Decision -> World Bridge creates CommunicationMessage, no downstream component
// may modify its core semantic content. Executive and World modules (Text, Voice, GUI, API,
// and future realization modules) may transform presentation only (such as grammar, formatting,
// wording, sentence order, punctuation, voice style, or rendering).
// They MUST NEVER modify:
//   - Intent
//   - DialogueAct
//   - Meaning
//   - Goal
//   - Slots
//   - Entities
//   - Confidence
// Semantic meaning must remain unchanged across all downstream transformations.
type CommunicationMessage struct {
	ResponseID string `json:"response_id"`
	ParentRef  string `json:"parent_ref"`

	// Core Semantic Content (Immutable across Executive and World layer)
	Intent      string  `json:"intent"`
	DialogueAct string  `json:"dialogue_act"`
	Meaning     string  `json:"meaning"`
	Goal        string  `json:"goal"`
	Confidence  float64 `json:"confidence"`

	// Presentation & Realization Hints (Can guide downstream rendering)
	Tone      string `json:"tone"`
	Verbosity string `json:"verbosity"`
	Language  string `json:"language"`
	Modality  string `json:"modality"`

	// Strongly Typed Semantic Grounding (Immutable)
	Slots    []Slot   `json:"slots"`
	Entities []Entity `json:"entities"`

	// Metadata represents diagnostic and extensibility information.
	//
	// Architectural Rule: Metadata Purpose
	// Metadata is intended ONLY for:
	//   - future extensibility
	//   - optional developer information
	//   - diagnostics
	//   - telemetry
	//   - debugging
	// Metadata MUST NOT become a dumping ground for required semantic information.
	// All core semantic information must remain strongly typed in the fields above.
	Metadata map[string]any `json:"metadata,omitempty"`

	// Optional developer/debugging lineage fields.
	// These exist ONLY for debugging, replay, telemetry, and diagnostics.
	// World modules and language generators MUST NEVER depend on these fields for normal operation.
	DecisionRef      string `json:"decision_ref,omitempty"`
	PlanRef          string `json:"plan_ref,omitempty"`
	ReasoningRef     string `json:"reasoning_ref,omitempty"`
	SemanticFrameRef string `json:"semantic_frame_ref,omitempty"`
}

// Validate checks whether CommunicationMessage is well-formed.
func (m *CommunicationMessage) Validate() error {
	if m == nil {
		return ErrNilMessage
	}
	if m.ResponseID == "" {
		return ErrMissingResponseID
	}
	return nil
}

// Marshal serializes CommunicationMessage to canonical JSON.
func Marshal(m *CommunicationMessage) ([]byte, error) {
	if m == nil {
		return nil, ErrNilMessage
	}
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("boundary validation failed: %w", err)
	}
	return json.Marshal(m)
}

// Unmarshal deserializes JSON bytes into a CommunicationMessage struct.
func Unmarshal(data []byte) (*CommunicationMessage, error) {
	if len(data) == 0 {
		return nil, ErrNilMessage
	}
	var m CommunicationMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("boundary: failed to unmarshal CommunicationMessage: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("boundary validation failed after unmarshal: %w", err)
	}
	return &m, nil
}
