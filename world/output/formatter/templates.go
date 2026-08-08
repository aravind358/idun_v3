package formatter

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"
	"time"
)

// Engine manages deterministic template realization.
type Engine struct {
	templateDir string
}

// NewEngine creates a new deterministic realization engine.
func NewEngine(templateDir string) *Engine {
	return &Engine{
		templateDir: templateDir,
	}
}

// Format executes the requested template using semantic data and presentation options.
func (e *Engine) Format(responseType string, data map[string]interface{}, operation string, err error) (string, error) {
	if responseType == "" {
		return "", fmt.Errorf("deterministic realization failed: empty response type")
	}

	funcMap := template.FuncMap{
		"formatTime": func(layout, value string) string {
			t, parseErr := time.Parse(time.RFC3339Nano, value)
			if parseErr != nil {
				t, parseErr = time.Parse(time.RFC3339, value)
				if parseErr != nil {
					return value
				}
			}
			return t.Format(layout)
		},
	}

	tmplName := responseType + ".tmpl"
	tmplPath := filepath.Join(e.templateDir, tmplName)

	// Read and parse template
	tmpl, parseErr := template.New(tmplName).Funcs(funcMap).ParseFiles(tmplPath)
	if parseErr != nil {
		return "", fmt.Errorf("failed to parse template for response type '%s': %w", responseType, parseErr)
	}

	viewData := make(map[string]interface{})
	if data != nil {
		for k, v := range data {
			viewData[k] = v
		}
	}
	
	viewData["operation"] = operation

	if err != nil {
		viewData["Error"] = err
	}

	var buf bytes.Buffer
	if execErr := tmpl.Execute(&buf, viewData); execErr != nil {
		return "", fmt.Errorf("failed to execute template: %w", execErr)
	}

	return buf.String(), nil
}
