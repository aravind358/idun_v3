package realization

import (
	"time"

	"idun/boundary"
)

// Tone classifies the desired surface presentation style without altering semantic truth.
type Tone string

const (
	ToneProfessional   Tone = "professional"
	ToneConversational Tone = "conversational"
	ToneConcise        Tone = "concise"
)

// ExecutionResponse is the legacy finalized artifact consumed by Language Realization.
// Under Phase 2 Communication Boundary Architecture, boundary.CommunicationMessage is the
// canonical input contract consumed by Language Realization and other World modules.
type ExecutionResponse struct {
	ResponseID       string `json:"response_id"`
	ParentRef        string `json:"parent_ref"`        // Pass-through routing header for World session correlation
	FinalizedContent string `json:"finalized_content"` // Complete, self-contained factual summary ready for realization
	Tone             Tone   `json:"tone"`              // Desired surface phrasing style
	Language         string `json:"language"`          // Target language/locale (e.g., "en-US", "ja-JP")
}

// ToCommunicationMessage converts a legacy ExecutionResponse to canonical boundary.CommunicationMessage.
func (e ExecutionResponse) ToCommunicationMessage() *boundary.CommunicationMessage {
	return &boundary.CommunicationMessage{
		ResponseID: e.ResponseID,
		ParentRef:  e.ParentRef,
		Meaning:    e.FinalizedContent,
		Goal:       e.FinalizedContent,
		Tone:       string(e.Tone),
		Language:   e.Language,
		Modality:   "text",
	}
}

// RealizedOutput is the minimal output contract published for World delivery.
type RealizedOutput struct {
	OutputID         string    `json:"output_id"`
	SourceResponseID string    `json:"source_response_id"` // Exact ID of the ExecutionResponse realized
	ParentRef        string    `json:"parent_ref"`         // Pass-through correlation ID for World session
	RealizedText     string    `json:"realized_text"`      // The polished natural human language string
	CreatedAt        time.Time `json:"created_at"`
}
