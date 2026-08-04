package deterministic

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"
	"time"

	"idun/capabilities"
	"idun/presentation"
)

// Engine implements the presentation.RealizationEngine for deterministic outputs.
type Engine struct {
	templateDir string
}

// NewEngine creates a new deterministic realization engine.
func NewEngine(templateDir string) *Engine {
	return &Engine{
		templateDir: templateDir,
	}
}

// Realize implements presentation.RealizationEngine.
func (e *Engine) Realize(ctx context.Context, res capabilities.CapabilityResult, parentRef string, responseID string) (*presentation.RealizedOutput, error) {
	if res.ResponseType == "" {
		return nil, fmt.Errorf("deterministic realization failed: empty ResponseType")
	}

	tmplPath := filepath.Join(e.templateDir, res.ResponseType+".tmpl")
	
	// Read and parse template
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template for response type '%s': %w", res.ResponseType, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, res.Data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return &presentation.RealizedOutput{
		OutputID:         "rlz-" + time.Now().UTC().Format("20060102150405.000000"),
		SourceResponseID: responseID,
		ParentRef:        parentRef,
		RealizedText:     buf.String(),
		CreatedAt:        time.Now().UTC(),
	}, nil
}
