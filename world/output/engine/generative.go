package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"idun/world/output"
)

// GenerativeInferenceService abstracts the LLM interface required for generative realization.
type GenerativeInferenceService interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// GenerativeRealizer uses a generative AI model to realize human-readable output.
type GenerativeRealizer struct {
	inferenceSvc GenerativeInferenceService
}

// NewGenerativeRealizer constructs a new GenerativeRealizer.
func NewGenerativeRealizer(svc GenerativeInferenceService) *GenerativeRealizer {
	return &GenerativeRealizer{
		inferenceSvc: svc,
	}
}

// Realize executes a generative LLM request based on the parsed data payload.
func (r *GenerativeRealizer) Realize(ctx context.Context, response output.CompositeResponse, desc output.Descriptor) (output.OutputDocument, error) {
	if len(response.ResolvedData) == 0 {
		return output.OutputDocument{}, fmt.Errorf("no resolved data to realize")
	}

	payload := response.ResolvedData[0]

	contentBytes, err := json.Marshal(payload.Data)
	var content string
	if err == nil {
		content = string(contentBytes)
	} else {
		content = fmt.Sprintf("%v", payload.Data)
	}

	promptStr := BuildRealizationPrompt(response.Options, content, nil)

	text, err := r.inferenceSvc.Generate(ctx, promptStr)
	if err != nil {
		return output.OutputDocument{}, fmt.Errorf("generative realization failed: %w", err)
	}

	return output.OutputDocument{
		ID:        "out_" + response.ExecutionID,
		Content:   text,
		CreatedAt: time.Now(),
	}, nil
}
