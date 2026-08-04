package presentation

import "time"

// RealizedOutput is the minimal output contract published for World delivery.
type RealizedOutput struct {
	OutputID         string    `json:"output_id"`
	SourceResponseID string    `json:"source_response_id"`
	ParentRef        string    `json:"parent_ref"`
	RealizedText     string    `json:"realized_text"`
	CreatedAt        time.Time `json:"created_at"`
}
