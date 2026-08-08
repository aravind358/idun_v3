package text

import (
	"context"
	"fmt"
	"io"

	"idun/world/output"
)

// OutputAdapter performs physical I/O for the Text plugin.
type OutputAdapter interface {
	Write(ctx context.Context, doc output.OutputDocument) error
}

// TextWriterAdapter implements OutputAdapter by writing OutputDocument content
// to any io.Writer (e.g. os.Stdout or a test bytes.Buffer).
// It is the canonical physical-I/O bridge between the World Output pipeline and
// the text-based external interface.
type TextWriterAdapter struct {
	writer io.Writer
}

// NewTextWriterAdapter creates a TextWriterAdapter that writes to w.
// w must not be nil.
func NewTextWriterAdapter(w io.Writer) *TextWriterAdapter {
	return &TextWriterAdapter{writer: w}
}

// Write formats the OutputDocument's Content and writes it to the underlying io.Writer.
func (a *TextWriterAdapter) Write(ctx context.Context, doc output.OutputDocument) error {
	if a.writer == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	_, err := fmt.Fprintf(a.writer, "%s\n", doc.Content)
	return err
}
