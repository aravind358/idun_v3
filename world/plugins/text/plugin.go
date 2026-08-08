package text

import (
	"context"

	"idun/world"
	"idun/world/output"
	"idun/world/output/formatter"
)

// Plugin implements output.OutputPlugin for the "text" modality.
type Plugin struct {
	fmt     formatter.Formatter
	adapter OutputAdapter
}

// NewPlugin creates a new Text OutputPlugin.
// fmt applies modality-specific formatting (may be nil for no-op formatting).
// adapter performs physical I/O (required for actual output; if nil, Write is a no-op).
func NewPlugin(fmt formatter.Formatter, adapter OutputAdapter) *Plugin {
	return &Plugin{
		fmt:     fmt,
		adapter: adapter,
	}
}

// Name returns the plugin's canonical name.
func (p *Plugin) Name() string {
	return "TextPlugin"
}

// SupportedModalities returns the modalities this plugin handles.
func (p *Plugin) SupportedModalities() []world.Modality {
	return []world.Modality{world.ModalityText}
}

// Priority returns the plugin's selection priority. Higher values take precedence.
func (p *Plugin) Priority() int {
	return 100
}

// Realize is part of the OutputPlugin interface.
// In the O4 architecture the OutputEngine performs realization globally before
// the plugin is called; the plugin's Realize method is a pass-through.
func (p *Plugin) Realize(ctx context.Context, response output.CompositeResponse) (output.OutputDocument, error) {
	return output.OutputDocument{}, nil
}

// Format applies modality-specific formatting to the OutputDocument.
func (p *Plugin) Format(ctx context.Context, doc output.OutputDocument) error {
	if p.fmt != nil {
		return p.fmt.Format(ctx, doc)
	}
	return nil
}

// Write performs physical I/O by delivering the OutputDocument via the adapter.
func (p *Plugin) Write(ctx context.Context, out any) error {
	if p.adapter == nil {
		return nil
	}
	doc, ok := out.(output.OutputDocument)
	if !ok {
		return nil
	}
	return p.adapter.Write(ctx, doc)
}
