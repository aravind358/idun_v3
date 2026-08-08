package text

import (
	"context"
	
	"idun/world/output"
)

// DefaultTextFormatter implements formatter.Formatter for the text plugin.
type DefaultTextFormatter struct{}

// NewDefaultTextFormatter creates a new DefaultTextFormatter.
func NewDefaultTextFormatter() *DefaultTextFormatter {
	return &DefaultTextFormatter{}
}

// Format applies modality-specific formatting to the OutputDocument.
func (f *DefaultTextFormatter) Format(ctx context.Context, doc output.OutputDocument) error {
	// For text, we might apply terminal colors or Markdown rendering here.
	// For now, it's a no-op as the OutputEngine provided raw text content.
	return nil
}
