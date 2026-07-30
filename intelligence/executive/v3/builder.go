package v3

import (
	"errors"
	"idun/core/foundation"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidParentArtifactID = errors.New("parent artifact ID is required")
	ErrInvalidEnvelopeID       = errors.New("envelope ID is required")
	ErrInvalidStatus           = errors.New("invalid execution status")
	ErrFailedWithoutReason     = errors.New("execution failed but no reason was provided")
	ErrInvalidNodeResult       = errors.New("node result is missing required fields")
)

// Builder ensures ExecutionResult is constructed immutably and perfectly validated.
type Builder struct {
	result *ExecutionResult
}

func NewBuilder() *Builder {
	return &Builder{
		result: &ExecutionResult{
			artifactID:  foundation.ArtifactID(uuid.New().String()),
			version:     "3.0",
			timestamp:   time.Now().UTC(),
			nodeResults: make(map[string]NodeResult),
		},
	}
}

func (b *Builder) WithParentArtifactID(id foundation.ArtifactID) *Builder {
	b.result.parentArtifactID = id
	return b
}

func (b *Builder) WithEnvelopeID(id string) *Builder {
	b.result.envelopeID = id
	return b
}

func (b *Builder) WithStatus(status ExecutionStatus) *Builder {
	b.result.status = status
	return b
}

func (b *Builder) WithOverallError(err string) *Builder {
	b.result.overallError = err
	return b
}

func (b *Builder) WithTotalDuration(d time.Duration) *Builder {
	b.result.totalDuration = d
	return b
}

func (b *Builder) AddNodeResult(res NodeResult) *Builder {
	b.result.nodeResults[res.NodeID] = res
	return b
}

func (b *Builder) Build() (*ExecutionResult, error) {
	if b.result.parentArtifactID == "" {
		return nil, ErrInvalidParentArtifactID
	}
	if b.result.envelopeID == "" {
		return nil, ErrInvalidEnvelopeID
	}

	switch b.result.status {
	case StatusCompleted, StatusTimedOut, StatusCancelled, StatusImpasse:
		// Valid
	case StatusFailed:
		if b.result.overallError == "" {
			return nil, ErrFailedWithoutReason
		}
	default:
		return nil, ErrInvalidStatus
	}

	for _, nr := range b.result.nodeResults {
		if nr.NodeID == "" {
			return nil, ErrInvalidNodeResult
		}
	}

	return b.result, nil
}
