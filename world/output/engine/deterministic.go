package engine

import (
	"context"
	"fmt"
	"time"

	"idun/world/output"
	"idun/world/output/formatter"
)

// DeterministicRealizer realizes capability payloads using deterministic templates.
type DeterministicRealizer struct {
	tmplEngine *formatter.Engine
}

// NewDeterministicRealizer constructs a new DeterministicRealizer.
func NewDeterministicRealizer(tmplEngine *formatter.Engine) *DeterministicRealizer {
	return &DeterministicRealizer{
		tmplEngine: tmplEngine,
	}
}

// Realize executes the deterministic template defined in the descriptor.
func (r *DeterministicRealizer) Realize(ctx context.Context, response output.CompositeResponse, desc output.Descriptor) (output.OutputDocument, error) {
	if len(response.ResolvedData) == 0 {
		return output.OutputDocument{}, fmt.Errorf("no resolved data to realize")
	}

	payload := response.ResolvedData[0]

	// Prepare template data from payload
	var viewData map[string]interface{}
	if dataMap, ok := payload.Data.(map[string]interface{}); ok {
		viewData = dataMap
	} else if payload.Data != nil {
		viewData = map[string]interface{}{
			"payload": payload.Data,
		}
	}

	// Format using the template engine
	// Note: formatter.Engine.Format expects (responseType, data, operation, err)
	// We pass the TemplateID as the "responseType" so it loads `<TemplateID>.tmpl`
	text, err := r.tmplEngine.Format(desc.TemplateID, viewData, "", nil)
	if err != nil {
		return output.OutputDocument{}, fmt.Errorf("deterministic realization failed: %w", err)
	}

	return output.OutputDocument{
		ID:        "out_" + response.ExecutionID,
		Content:   text,
		CreatedAt: time.Now(),
	}, nil
}
