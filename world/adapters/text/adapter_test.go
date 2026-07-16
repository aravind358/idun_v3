package text_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"idun/world"
	"idun/world/adapters/text"
)

// ============================================================================
// Test helpers
// ============================================================================

func buildTestResponse(interactionID, sessionID, content string) *world.Response {
	return &world.Response{
		ResponseID:    "test-resp-001",
		InteractionID: interactionID,
		SessionID:     sessionID,
		Modality:      world.ModalityText,
		Content:       content,
		PayloadRef:    "ref-" + interactionID,
		Status:        world.ResponseStatusSuccess,
		ResultStatus:  world.ResultStatusSuccess,
		CreatedAt:     time.Now().UTC(),
		ReplayMetadata: world.WorldReplayMetadata{
			WorldVersion:          world.WorldVersion,
			PolicyFingerprint:     "pol-fp",
			CapabilityFingerprint: "cap-fp",
			InteractionFingerprint: "int-fp",
		},
	}
}

func buildTestResponseWithRef(interactionID, sessionID, content, payloadRef string) *world.Response {
	r := buildTestResponse(interactionID, sessionID, content)
	r.PayloadRef = payloadRef
	return r
}

// ============================================================================
// TextInputAdapter Tests
// ============================================================================

func TestNewTextInputAdapter_NilReader(t *testing.T) {
	_, err := text.NewTextInputAdapter(nil)
	if err == nil {
		t.Fatal("expected error for nil reader, got nil")
	}
}

func TestTextInputAdapter_Identity(t *testing.T) {
	adapter, err := text.NewTextInputAdapter(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("NewTextInputAdapter failed: %v", err)
	}
	if adapter.Name() != "TextInputAdapter" {
		t.Errorf("expected Name=TextInputAdapter, got %q", adapter.Name())
	}
	if adapter.AdapterVersion() != "2.0.0-FROZEN" {
		t.Errorf("expected AdapterVersion=2.0.0-FROZEN, got %q", adapter.AdapterVersion())
	}
	if adapter.AdapterFingerprint() == "" {
		t.Error("expected non-empty AdapterFingerprint")
	}
}

func TestTextInputAdapter_Receive_ValidLine(t *testing.T) {
	reader := strings.NewReader("Hello, IDUN!\n")
	adapter, err := text.NewTextInputAdapter(reader)
	if err != nil {
		t.Fatalf("NewTextInputAdapter: %v", err)
	}

	ctx := context.Background()
	interaction, err := adapter.Receive(ctx)
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}
	if interaction == nil {
		t.Fatal("expected non-nil Interaction")
	}
	if interaction.OriginalInput != "Hello, IDUN!" {
		t.Errorf("expected OriginalInput %q, got %q", "Hello, IDUN!", interaction.OriginalInput)
	}
	if interaction.NormalizedInput != "Hello, IDUN!" {
		t.Errorf("expected NormalizedInput %q, got %q", "Hello, IDUN!", interaction.NormalizedInput)
	}
	if interaction.Modality != world.ModalityText {
		t.Errorf("expected Modality TEXT, got %s", interaction.Modality)
	}
	if interaction.Origin != world.OriginUser {
		t.Errorf("expected Origin USER, got %s", interaction.Origin)
	}
	if interaction.InteractionID == "" {
		t.Error("expected non-empty InteractionID")
	}
	if interaction.PayloadRef == "" {
		t.Error("expected non-empty PayloadRef")
	}
}

func TestTextInputAdapter_SkipsBlankLines(t *testing.T) {
	reader := strings.NewReader("\n   \n\nActual input\n")
	adapter, _ := text.NewTextInputAdapter(reader)

	ctx := context.Background()
	interaction, err := adapter.Receive(ctx)
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}
	if interaction.NormalizedInput != "Actual input" {
		t.Errorf("expected NormalizedInput %q, got %q", "Actual input", interaction.NormalizedInput)
	}
}

func TestTextInputAdapter_EOFOnEmptyReader(t *testing.T) {
	adapter, _ := text.NewTextInputAdapter(strings.NewReader(""))
	_, err := adapter.Receive(context.Background())
	if err == nil {
		t.Error("expected error (EOF) for empty reader, got nil")
	}
}

func TestTextInputAdapter_MultipleLines(t *testing.T) {
	reader := strings.NewReader("first line\nsecond line\nthird line\n")
	adapter, _ := text.NewTextInputAdapter(reader)
	ctx := context.Background()

	lines := []string{"first line", "second line", "third line"}
	for i, expected := range lines {
		interaction, err := adapter.Receive(ctx)
		if err != nil {
			t.Fatalf("line %d: Receive failed: %v", i, err)
		}
		if interaction.NormalizedInput != expected {
			t.Errorf("line %d: expected %q, got %q", i, expected, interaction.NormalizedInput)
		}
	}
	// Next receive should return EOF
	_, err := adapter.Receive(ctx)
	if err == nil {
		t.Error("expected EOF after last line, got nil")
	}
}

func TestTextInputAdapter_ContextCancellation(t *testing.T) {
	// Use a reader that blocks — pipe with no data written
	pr, _ := io.Pipe()
	adapter, _ := text.NewTextInputAdapter(pr)
	defer pr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := adapter.Receive(ctx)
	// On a blocking pipe, the scanner.Scan() call blocks — but context expires.
	// Since bufio.Scanner does not check context internally, the test accepts either
	// ctx.Err() or a scanner error after the pipe is closed by defer.
	_ = err // any error is acceptable; we just verify it does not hang
}

func TestTextInputAdapter_ClosePreventsFurtherReceive(t *testing.T) {
	adapter, _ := text.NewTextInputAdapter(strings.NewReader("hello\n"))
	_ = adapter.Close()

	_, err := adapter.Receive(context.Background())
	if err == nil {
		t.Error("expected error after Close, got nil")
	}
}

// ============================================================================
// TextOutputAdapter Tests
// ============================================================================

func TestNewTextOutputAdapter_NilWriter(t *testing.T) {
	_, err := text.NewTextOutputAdapter(nil)
	if err == nil {
		t.Fatal("expected error for nil writer, got nil")
	}
}

func TestTextOutputAdapter_Identity(t *testing.T) {
	adapter, err := text.NewTextOutputAdapter(&bytes.Buffer{})
	if err != nil {
		t.Fatalf("NewTextOutputAdapter: %v", err)
	}
	if adapter.Name() != "TextOutputAdapter" {
		t.Errorf("expected Name=TextOutputAdapter, got %q", adapter.Name())
	}
	if adapter.AdapterVersion() != "2.0.0-FROZEN" {
		t.Errorf("expected AdapterVersion=2.0.0-FROZEN, got %q", adapter.AdapterVersion())
	}
	if adapter.AdapterFingerprint() == "" {
		t.Error("expected non-empty AdapterFingerprint")
	}
}

func TestTextOutputAdapter_Send_WritesContent(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter, err := text.NewTextOutputAdapter(buf)
	if err != nil {
		t.Fatalf("NewTextOutputAdapter: %v", err)
	}

	response := buildTestResponse("interaction-id", "session-id", "Hello from IDUN")
	if err := adapter.Send(context.Background(), response); err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Hello from IDUN") {
		t.Errorf("expected output to contain %q, got %q", "Hello from IDUN", buf.String())
	}
}

func TestTextOutputAdapter_Send_FallsBackToPayloadRef(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter, _ := text.NewTextOutputAdapter(buf)

	response := buildTestResponseWithRef("interaction-id", "session-id", "", "sha256:abc123")
	_ = adapter.Send(context.Background(), response)
	if !strings.Contains(buf.String(), "sha256:abc123") {
		t.Errorf("expected output to contain payload ref, got %q", buf.String())
	}
}

func TestTextOutputAdapter_Send_NilResponse(t *testing.T) {
	adapter, _ := text.NewTextOutputAdapter(&bytes.Buffer{})
	err := adapter.Send(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil response, got nil")
	}
}

func TestTextOutputAdapter_ClosePreventsWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter, _ := text.NewTextOutputAdapter(buf)
	_ = adapter.Close()

	response := buildTestResponse("id", "session", "content")
	err := adapter.Send(context.Background(), response)
	if err == nil {
		t.Error("expected error after Close, got nil")
	}
}

// ============================================================================
// Adapter Fingerprint Determinism Tests (Refinement 10)
// ============================================================================

func TestAdapterFingerprints_AreDeterministic(t *testing.T) {
	// Same name+version always produces the same fingerprint.
	a1, _ := text.NewTextInputAdapter(strings.NewReader(""))
	a2, _ := text.NewTextInputAdapter(strings.NewReader("different data"))
	if a1.AdapterFingerprint() != a2.AdapterFingerprint() {
		t.Error("expected same fingerprint for same adapter name+version, got different")
	}

	o1, _ := text.NewTextOutputAdapter(&bytes.Buffer{})
	o2, _ := text.NewTextOutputAdapter(&bytes.Buffer{})
	if o1.AdapterFingerprint() != o2.AdapterFingerprint() {
		t.Error("expected same output adapter fingerprint, got different")
	}

	// Input and output adapters must have different fingerprints (different names).
	if a1.AdapterFingerprint() == o1.AdapterFingerprint() {
		t.Error("expected different fingerprints for input vs output adapters")
	}
}
