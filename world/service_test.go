package world_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/workspace"
	"idun/world"
	"idun/world/adapters/text"
)

// ============================================================================
// Test payload storer (minimal mock)
// ============================================================================

type mockPayloadStorer struct {
	stored map[string][]byte
}

func newMockPayloadStorer() *mockPayloadStorer {
	return &mockPayloadStorer{stored: make(map[string][]byte)}
}

func (m *mockPayloadStorer) Store(_ context.Context, data []byte) (string, error) {
	key := "mock-ref-" + string(data[:min(8, len(data))])
	m.stored[key] = data
	return key, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// NewService Construction Tests
// ============================================================================

func TestNewService_NilWorkspace(t *testing.T) {
	input, _ := text.NewTextInputAdapter(bytes.NewReader([]byte("hello\n")))
	output, _ := text.NewTextOutputAdapter(&bytes.Buffer{})
	storer := newMockPayloadStorer()

	_, err := world.NewService(nil, input, output, storer)
	if err == nil {
		t.Fatal("expected error for nil workspace, got nil")
	}
}

func TestNewService_NilInputAdapter(t *testing.T) {
	ws := workspace.NewEngine()
	output, _ := text.NewTextOutputAdapter(&bytes.Buffer{})
	storer := newMockPayloadStorer()

	_, err := world.NewService(ws, nil, output, storer)
	if err == nil {
		t.Fatal("expected error for nil input adapter, got nil")
	}
}

func TestNewService_NilOutputAdapter(t *testing.T) {
	ws := workspace.NewEngine()
	input, _ := text.NewTextInputAdapter(bytes.NewReader([]byte("hello\n")))
	storer := newMockPayloadStorer()

	_, err := world.NewService(ws, input, nil, storer)
	if err == nil {
		t.Fatal("expected error for nil output adapter, got nil")
	}
}

func TestNewService_NilPayloadStorer(t *testing.T) {
	ws := workspace.NewEngine()
	input, _ := text.NewTextInputAdapter(bytes.NewReader([]byte("hello\n")))
	output, _ := text.NewTextOutputAdapter(&bytes.Buffer{})

	_, err := world.NewService(ws, input, output, nil)
	if err == nil {
		t.Fatal("expected error for nil payload storer, got nil")
	}
}

// ============================================================================
// Service Lifecycle Tests
// ============================================================================

func buildTestService(t *testing.T) (*world.Service, *bytes.Buffer) {
	t.Helper()
	ws := workspace.NewEngine()
	if err := ws.Start(); err != nil {
		t.Fatalf("workspace.Start failed: %v", err)
	}
	t.Cleanup(func() { _ = ws.Close() })

	input, _ := text.NewTextInputAdapter(bytes.NewReader([]byte("hello world\n")))
	outBuf := &bytes.Buffer{}
	output, _ := text.NewTextOutputAdapter(outBuf)
	storer := newMockPayloadStorer()

	svc, err := world.NewService(ws, input, output, storer)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}
	return svc, outBuf
}

func TestService_Name(t *testing.T) {
	svc, _ := buildTestService(t)
	if svc.Name() != "World.Service" {
		t.Errorf("expected Name=World.Service, got %s", svc.Name())
	}
}

func TestService_GetPolicyProfile(t *testing.T) {
	svc, _ := buildTestService(t)
	p := svc.GetPolicyProfile()
	if p == nil {
		t.Fatal("expected non-nil policy profile")
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("policy profile Validate failed: %v", err)
	}
}

func TestService_GetCapabilities(t *testing.T) {
	svc, _ := buildTestService(t)
	c := svc.GetCapabilities()
	if c == nil {
		t.Fatal("expected non-nil capabilities")
	}
	if !c.SupportsText {
		t.Error("expected SupportsText=true")
	}
}

func TestService_Lifecycle_StartClose(t *testing.T) {
	svc, _ := buildTestService(t)
	ctx := context.Background()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	// Second Start is idempotent
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("second Start failed: %v", err)
	}
	if err := svc.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	// Second Close is idempotent
	if err := svc.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}
}

func TestService_HandleInteraction_BeforeStart(t *testing.T) {
	svc, _ := buildTestService(t)

	interaction := buildValidTestInteraction()
	err := svc.HandleInteraction(context.Background(), interaction)
	if err == nil {
		t.Fatal("expected error for HandleInteraction before Start, got nil")
	}
}

func TestService_HandleInteraction_AfterClose(t *testing.T) {
	svc, _ := buildTestService(t)
	ctx := context.Background()
	_ = svc.Start(ctx)
	_ = svc.Close()

	interaction := buildValidTestInteraction()
	err := svc.HandleInteraction(ctx, interaction)
	if err == nil {
		t.Fatal("expected error for HandleInteraction after Close, got nil")
	}
}

// ============================================================================
// HandleInteraction Event-Driven Tests (Refinement 11)
// ============================================================================

func TestService_HandleInteraction_PublishesToWorkspace(t *testing.T) {
	ws := workspace.NewEngine()
	if err := ws.Start(); err != nil {
		t.Fatalf("workspace.Start failed: %v", err)
	}
	defer ws.Close()

	input, _ := text.NewTextInputAdapter(bytes.NewReader([]byte("hello\n")))
	outBuf := &bytes.Buffer{}
	output, _ := text.NewTextOutputAdapter(outBuf)
	storer := newMockPayloadStorer()

	svc, err := world.NewService(ws, input, output, storer)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer svc.Close()

	// Subscribe to TopicPerception to verify World publishes there (event-driven, Refinement 11).
	received := make(chan struct{}, 1)
	_, err = ws.Subscribe("perception", "test-observer", func(_ context.Context, env communication.Envelope) error {
		if env.Source == "World.Service" {
			select {
			case received <- struct{}{}:
			default:
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	interaction := buildValidTestInteraction()
	if err := svc.HandleInteraction(ctx, interaction); err != nil {
		t.Fatalf("HandleInteraction failed: %v", err)
	}

	// Verify publication happens (allow brief async processing time).
	select {
	case <-received:
		// World successfully published to TopicPerception — event-driven flow confirmed.
	case <-time.After(200 * time.Millisecond):
		t.Error("expected World.Service to publish to TopicPerception within 200ms")
	}
}

func TestService_HandleInteraction_NilInteraction(t *testing.T) {
	svc, _ := buildTestService(t)
	ctx := context.Background()
	_ = svc.Start(ctx)
	defer svc.Close()

	err := svc.HandleInteraction(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil interaction, got nil")
	}
}

func TestService_HandleInteraction_InvalidInteraction(t *testing.T) {
	svc, _ := buildTestService(t)
	ctx := context.Background()
	_ = svc.Start(ctx)
	defer svc.Close()

	// Invalid — missing SessionID
	interaction := &world.Interaction{
		InteractionID: "id",
		Origin:        world.OriginUser,
		Modality:      world.ModalityText,
		OriginalInput: "hello",
		PayloadRef:    "ref",
	}
	err := svc.HandleInteraction(ctx, interaction)
	if err == nil {
		t.Fatal("expected error for invalid interaction (missing SessionID), got nil")
	}
}

// ============================================================================
// CreateInteraction Tests
// ============================================================================

func TestService_CreateInteraction_Valid(t *testing.T) {
	svc, _ := buildTestService(t)
	ctx := context.Background()

	interaction, err := svc.CreateInteraction(ctx, "Hello, IDUN!", "session-001", world.OriginUser, world.ModalityText)
	if err != nil {
		t.Fatalf("CreateInteraction failed: %v", err)
	}
	if interaction == nil {
		t.Fatal("expected non-nil interaction")
	}
	if interaction.OriginalInput != "Hello, IDUN!" {
		t.Errorf("expected OriginalInput=Hello, IDUN!, got %q", interaction.OriginalInput)
	}
	if interaction.NormalizedInput == "" {
		t.Error("expected non-empty NormalizedInput")
	}
	if interaction.ReplayMetadata.InteractionFingerprint == "" {
		t.Error("expected non-empty InteractionFingerprint in ReplayMetadata")
	}
	if interaction.ReplayMetadata.WorldVersion != world.WorldVersion {
		t.Errorf("expected WorldVersion=%s, got %s", world.WorldVersion, interaction.ReplayMetadata.WorldVersion)
	}
	if err := interaction.Validate(); err != nil {
		t.Fatalf("CreateInteraction produced invalid Interaction: %v", err)
	}
}

func TestService_CreateInteraction_EmptyInput_Rejected(t *testing.T) {
	svc, _ := buildTestService(t)
	ctx := context.Background()

	_, err := svc.CreateInteraction(ctx, "   ", "session-001", world.OriginUser, world.ModalityText)
	if err == nil {
		t.Fatal("expected error for empty input (after normalization), got nil")
	}
}

func TestService_CreateInteraction_AppliesDefaultSession(t *testing.T) {
	svc, _ := buildTestService(t)
	ctx := context.Background()

	interaction, err := svc.CreateInteraction(ctx, "Hello", "", world.OriginUser, world.ModalityText)
	if err != nil {
		t.Fatalf("CreateInteraction failed: %v", err)
	}
	// When sessionID is empty, service should apply DefaultSessionID
	if interaction.SessionID == "" {
		t.Error("expected non-empty SessionID after applying DefaultSessionID")
	}
}

// ============================================================================
// WorldSummary Tests
// ============================================================================

func TestService_GetSummary_InitialZero(t *testing.T) {
	svc, _ := buildTestService(t)
	s := svc.GetSummary()
	if err := s.Validate(); err != nil {
		t.Fatalf("initial summary invalid: %v", err)
	}
	if s.TotalInteractions != 0 {
		t.Errorf("expected TotalInteractions=0, got %d", s.TotalInteractions)
	}
}

func TestService_GetSummary_AfterHandleInteraction(t *testing.T) {
	svc, _ := buildTestService(t)
	ctx := context.Background()
	_ = svc.Start(ctx)
	defer svc.Close()

	interaction, err := svc.CreateInteraction(ctx, "Hello", "session", world.OriginUser, world.ModalityText)
	if err != nil {
		t.Fatalf("CreateInteraction: %v", err)
	}
	_ = svc.HandleInteraction(ctx, interaction)

	// Allow async goroutine scheduling
	time.Sleep(10 * time.Millisecond)

	s := svc.GetSummary()
	if s.TotalInteractions < 1 {
		t.Errorf("expected TotalInteractions>=1, got %d", s.TotalInteractions)
	}
}

// ============================================================================
// Custom Policy and Capabilities Options Tests
// ============================================================================

func TestService_CustomPolicy(t *testing.T) {
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	input, _ := text.NewTextInputAdapter(bytes.NewReader([]byte{}))
	output, _ := text.NewTextOutputAdapter(&bytes.Buffer{})
	storer := newMockPayloadStorer()

	customPolicy, err := world.NewWorldPolicyProfileBuilder().
		WithMaximumInputLength(100).
		Build()
	if err != nil {
		t.Fatalf("policy build failed: %v", err)
	}

	svc, err := world.NewService(ws, input, output, storer,
		world.WithPolicy(customPolicy),
	)
	if err != nil {
		t.Fatalf("NewService with custom policy: %v", err)
	}

	p := svc.GetPolicyProfile()
	if p.MaximumInputLength != 100 {
		t.Errorf("expected MaximumInputLength=100, got %d", p.MaximumInputLength)
	}
}

func TestService_ResponseCorrelation_SuccessAndMismatch(t *testing.T) {
	ws := workspace.NewEngine()
	if err := ws.Start(); err != nil {
		t.Fatalf("workspace.Start failed: %v", err)
	}
	defer ws.Close()

	input, _ := text.NewTextInputAdapter(bytes.NewReader([]byte("hello\n")))
	outBuf := &bytes.Buffer{}
	output, _ := text.NewTextOutputAdapter(outBuf)
	storer := newMockPayloadStorer()

	svc, err := world.NewService(ws, input, output, storer)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer svc.Close()

	interaction := buildValidTestInteraction()
	if err := svc.HandleInteraction(ctx, interaction); err != nil {
		t.Fatalf("HandleInteraction failed: %v", err)
	}

	// 1. Publish an envelope on TopicActionExecution with a mismatched/empty ParentRef.
	// World must ignore it and not output anything.
	mismatchEnv, _ := communication.NewEnvelopeBuilder().
		WithSource("Intelligence.Executive").
		WithTopic(communication.TopicActionExecution).
		WithParentRef("mismatched-parent-id").
		WithPayloadRef("should not be printed").
		WithModality("structured-frame").
		WithConfidence(0.95).
		Build()
	if err := ws.Publish(ctx, mismatchEnv); err != nil {
		t.Fatalf("Publish mismatchEnv failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if outBuf.Len() > 0 {
		t.Fatalf("expected no output for mismatched ParentRef, got: %q", outBuf.String())
	}

	// 2. Publish an envelope on TopicActionExecution with the correct ParentRef (interaction.InteractionID).
	// World must correlate it and deliver the Response through output.Send.
	matchEnv, _ := communication.NewEnvelopeBuilder().
		WithSource("Intelligence.Executive").
		WithTopic(communication.TopicActionExecution).
		WithParentRef(interaction.InteractionID).
		WithPayloadRef("Hello from IDUN cognitive pipeline!").
		WithModality("structured-frame").
		WithConfidence(0.95).
		Build()
	if err := ws.Publish(ctx, matchEnv); err != nil {
		t.Fatalf("Publish matchEnv failed: %v", err)
	}

	// Wait up to 200ms for async response correlation and output delivery.
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if strings.Contains(outBuf.String(), "Hello from IDUN cognitive pipeline!") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !strings.Contains(outBuf.String(), "Hello from IDUN cognitive pipeline!") {
		t.Fatalf("expected output to contain response text via ParentRef correlation, got: %q", outBuf.String())
	}
}

// ============================================================================
// Helpers
// ============================================================================

func buildValidTestInteraction() *world.Interaction {
	return &world.Interaction{
		InteractionID:   "test-id-001",
		SessionID:       "session-001",
		Origin:          world.OriginUser,
		Modality:        world.ModalityText,
		OriginalInput:   "Hello, IDUN!",
		NormalizedInput: "Hello, IDUN!",
		PayloadRef:      "sha256:test-payload",
		CreatedAt:       time.Now().UTC(),
		ReplayMetadata: world.WorldReplayMetadata{
			WorldVersion:           world.WorldVersion,
			PolicyFingerprint:      "policy-fp",
			CapabilityFingerprint:  "cap-fp",
			InteractionFingerprint: "int-fp",
		},
	}
}
