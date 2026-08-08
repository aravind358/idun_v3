package output

import (
	"time"
)

// OutputDocument represents a structured, modality-agnostic response payload.
// It is produced by the OutputEngine and consumed by modality-specific plugins.
type OutputDocument struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	CreatedAt time.Time              `json:"created_at"`
	
	// Primary textual or spoken content
	Content   string                 `json:"content"`
	
	// Metadata contains formatting hints or structured data (e.g. JSON cards)
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
