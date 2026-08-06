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
func (e *Engine) Realize(ctx context.Context, res capabilities.CapabilityResult, pctx presentation.PresentationContext, responseID string) (*presentation.RealizedOutput, error) {
	if pctx.ResponseType == "" {
		return nil, fmt.Errorf("deterministic realization failed: empty ResponseType in context")
	}

	funcMap := template.FuncMap{
		"formatTime": func(layout, value string) string {
			t, err := time.Parse(time.RFC3339Nano, value)
			if err != nil {
				t, err = time.Parse(time.RFC3339, value)
				if err != nil {
					return value
				}
			}
			return t.Format(layout)
		},
	}

	tmplName := pctx.ResponseType + ".tmpl"
	tmplPath := filepath.Join(e.templateDir, tmplName)
	
	// Read and parse template
	tmpl, err := template.New(tmplName).Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template for response type '%s': %w", pctx.ResponseType, err)
	}

	// Merge pure semantic data and presentation hints/operations
	viewData := make(map[string]interface{})
	if res.Data != nil {
		for k, v := range res.Data {
			viewData[k] = v
		}
	}
	if pctx.PresentationHints != nil {
		for k, v := range pctx.PresentationHints {
			viewData[k] = v
		}
	}
	viewData["operation"] = pctx.Operation
	
	if res.Error != nil {
		viewData["Error"] = res.Error
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, viewData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return &presentation.RealizedOutput{
		OutputID:         "rlz-" + time.Now().UTC().Format("20060102150405.000000"),
		SourceResponseID: responseID,
		ParentRef:        pctx.ParentRef,
		RealizedText:     buf.String(),
		CreatedAt:        time.Now().UTC(),
	}, nil
}
