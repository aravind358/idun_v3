package v3

import (
	"encoding/json"
	"idun/core/foundation"
	"time"
)

type executionResultJSON struct {
	ArtifactID       foundation.ArtifactID `json:"artifact_id"`
	ParentArtifactID foundation.ArtifactID `json:"parent_artifact_id"`
	EnvelopeID       string                `json:"envelope_id"`
	Timestamp        time.Time             `json:"timestamp"`
	Version          string                `json:"version"`
	Status           ExecutionStatus       `json:"status"`
	NodeResults      map[string]NodeResult `json:"node_results"`
	TotalDuration    time.Duration         `json:"total_duration"`
	OverallError     string                `json:"overall_error,omitempty"`
}

// MarshalJSON implements custom JSON serialization.
func (e *ExecutionResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(executionResultJSON{
		ArtifactID:       e.artifactID,
		ParentArtifactID: e.parentArtifactID,
		EnvelopeID:       e.envelopeID,
		Timestamp:        e.timestamp,
		Version:          e.version,
		Status:           e.status,
		NodeResults:      e.nodeResults,
		TotalDuration:    e.totalDuration,
		OverallError:     e.overallError,
	})
}

// UnmarshalJSON implements custom JSON deserialization.
func (e *ExecutionResult) UnmarshalJSON(data []byte) error {
	var mirror executionResultJSON
	if err := json.Unmarshal(data, &mirror); err != nil {
		return err
	}
	e.artifactID = mirror.ArtifactID
	e.parentArtifactID = mirror.ParentArtifactID
	e.envelopeID = mirror.EnvelopeID
	e.timestamp = mirror.Timestamp
	e.version = mirror.Version
	e.status = mirror.Status
	e.nodeResults = mirror.NodeResults
	e.totalDuration = mirror.TotalDuration
	e.overallError = mirror.OverallError
	return nil
}
