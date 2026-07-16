package text

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"idun/world"
)

// adapterVersion is the canonical implementation version of the Text adapters.
const adapterVersion = "2.0.0-FROZEN"

// computeAdapterFingerprint derives a deterministic SHA-256 identity digest
// for an adapter from its name and version. This fingerprint changes only
// when the adapter name or version changes, enabling exact replay provenance.
func computeAdapterFingerprint(name, version string) string {
	raw := fmt.Sprintf("world-adapter|name:%s|version:%s", name, version)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// generateAdapterID returns a secure random 16-byte hex identifier.
func generateAdapterID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// ============================================================================
// TextInputAdapter
// ============================================================================

const textInputAdapterName = "TextInputAdapter"

// TextInputAdapter reads line-by-line text from any io.Reader (e.g. os.Stdin).
// Each non-empty line triggers an Interaction to be returned.
// The adapter carries immutable identity for deterministic replay provenance (Refinement 10).
type TextInputAdapter struct {
	reader             io.Reader
	scanner            *bufio.Scanner
	adapterFingerprint string
	closed             atomic.Bool
	linesCh            chan string
	errCh              chan error
	stopCh             chan struct{}
}

// NewTextInputAdapter constructs a new TextInputAdapter reading from the provided reader.
// If reader is nil, an error is returned — callers must inject os.Stdin explicitly.
func NewTextInputAdapter(reader io.Reader) (*TextInputAdapter, error) {
	if reader == nil {
		return nil, fmt.Errorf("world/adapters/text: TextInputAdapter reader cannot be nil")
	}
	a := &TextInputAdapter{
		reader:             reader,
		scanner:            bufio.NewScanner(reader),
		adapterFingerprint: computeAdapterFingerprint(textInputAdapterName, adapterVersion),
		linesCh:            make(chan string),
		errCh:              make(chan error, 1),
		stopCh:             make(chan struct{}),
	}
	go a.scanLoop()
	return a, nil
}

func (a *TextInputAdapter) scanLoop() {
	for a.scanner.Scan() {
		line := a.scanner.Text()
		select {
		case a.linesCh <- line:
		case <-a.stopCh:
			return
		}
	}
	if err := a.scanner.Err(); err != nil {
		select {
		case a.errCh <- err:
		case <-a.stopCh:
		}
	}
	close(a.linesCh)
}

// Name returns the canonical adapter name.
func (a *TextInputAdapter) Name() string {
	return textInputAdapterName
}

// AdapterVersion returns the immutable implementation version of this adapter.
func (a *TextInputAdapter) AdapterVersion() string {
	return adapterVersion
}

// AdapterFingerprint returns the deterministic SHA-256 identity of this adapter.
func (a *TextInputAdapter) AdapterFingerprint() string {
	return a.adapterFingerprint
}

// Receive blocks until a non-empty line is available from the underlying reader.
// Context cancellation causes Receive to return ctx.Err().
//
// The returned Interaction contains the raw OriginalInput exactly as received,
// and a trimmed NormalizedInput. Full policy normalization and fingerprinting
// is applied by Service.CreateInteraction before Workspace publication.
func (a *TextInputAdapter) Receive(ctx context.Context) (*world.Interaction, error) {
	if a.closed.Load() {
		return nil, fmt.Errorf("world/adapters/text: TextInputAdapter is closed")
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-a.errCh:
			return nil, fmt.Errorf("world/adapters/text: scanner error: %w", err)
		case line, ok := <-a.linesCh:
			if !ok {
				select {
				case err := <-a.errCh:
					return nil, fmt.Errorf("world/adapters/text: scanner error: %w", err)
				default:
					return nil, io.EOF
				}
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				// Skip blank lines — policy enforcement is in Service.CreateInteraction.
				continue
			}

			// Return a minimal Interaction. The World Service completes it via CreateInteraction.
			// We provide a minimal PayloadRef so the struct can pass standalone validation
			// in tests that do not go through the full service pipeline.
			return &world.Interaction{
				InteractionID:   generateAdapterID(),
				SessionID:       "text-session",
				Origin:          world.OriginUser,
				Modality:        world.ModalityText,
				OriginalInput:   line,
				NormalizedInput: trimmed,
				PayloadRef:      computeAdapterFingerprint(line, adapterVersion),
			}, nil
		}
	}
}

// Close marks the TextInputAdapter as closed. Subsequent Receive calls return an error.
func (a *TextInputAdapter) Close() error {
	if !a.closed.CompareAndSwap(false, true) {
		return nil
	}
	close(a.stopCh)
	if closer, ok := a.reader.(io.Closer); ok {
		_ = closer.Close()
	}
	return nil
}

// ============================================================================
// TextOutputAdapter
// ============================================================================

const textOutputAdapterName = "TextOutputAdapter"

// TextOutputAdapter writes Response content as formatted text to any io.Writer (e.g. os.Stdout).
// The adapter carries immutable identity for deterministic replay provenance (Refinement 10).
type TextOutputAdapter struct {
	writer             io.Writer
	adapterFingerprint string
	closed             atomic.Bool
}

// NewTextOutputAdapter constructs a new TextOutputAdapter writing to the provided writer.
// If writer is nil, an error is returned — callers must inject os.Stdout explicitly.
func NewTextOutputAdapter(writer io.Writer) (*TextOutputAdapter, error) {
	if writer == nil {
		return nil, fmt.Errorf("world/adapters/text: TextOutputAdapter writer cannot be nil")
	}
	return &TextOutputAdapter{
		writer:             writer,
		adapterFingerprint: computeAdapterFingerprint(textOutputAdapterName, adapterVersion),
	}, nil
}

// Name returns the canonical adapter name.
func (a *TextOutputAdapter) Name() string {
	return textOutputAdapterName
}

// AdapterVersion returns the immutable implementation version of this adapter.
func (a *TextOutputAdapter) AdapterVersion() string {
	return adapterVersion
}

// AdapterFingerprint returns the deterministic SHA-256 identity of this adapter.
func (a *TextOutputAdapter) AdapterFingerprint() string {
	return a.adapterFingerprint
}

// Send formats a Response and writes it to the underlying writer.
// World is content-blind: it presents Content as-is without interpretation.
// If the Response has no Content but has a PayloadRef, the ref is presented instead.
func (a *TextOutputAdapter) Send(ctx context.Context, response *world.Response) error {
	if a.closed.Load() {
		return fmt.Errorf("world/adapters/text: TextOutputAdapter is closed")
	}
	if response == nil {
		return fmt.Errorf("world/adapters/text: response cannot be nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	content := response.Content
	if content == "" && response.PayloadRef != "" {
		content = fmt.Sprintf("[ref:%s]", response.PayloadRef)
	}

	_, err := fmt.Fprintf(a.writer, "%s\n", content)
	return err
}

// Close marks the TextOutputAdapter as closed. Subsequent Send calls return an error.
func (a *TextOutputAdapter) Close() error {
	a.closed.Store(true)
	return nil
}
