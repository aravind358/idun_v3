package world

import (
	"testing"
	"time"
)

// ============================================================================
// WorldPolicyProfile Tests
// ============================================================================

func TestDefaultWorldPolicyProfile_Valid(t *testing.T) {
	p := DefaultWorldPolicyProfile()
	if err := p.Validate(); err != nil {
		t.Fatalf("DefaultWorldPolicyProfile.Validate failed: %v", err)
	}
	if p.PolicyFingerprint == "" {
		t.Error("expected non-empty PolicyFingerprint")
	}
	if p.MaximumInputLength <= 0 {
		t.Error("expected positive MaximumInputLength")
	}
}

func TestWorldPolicyProfile_Validate_MissingVersion(t *testing.T) {
	p := DefaultWorldPolicyProfile()
	p.PolicyVersion = ""
	p.PolicyFingerprint = ComputePolicyFingerprint(p)
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing PolicyVersion, got nil")
	}
}

func TestWorldPolicyProfile_Validate_ZeroInputLength(t *testing.T) {
	p := DefaultWorldPolicyProfile()
	p.MaximumInputLength = 0
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for zero MaximumInputLength, got nil")
	}
}

func TestWorldPolicyProfile_Validate_NegativeTimeout(t *testing.T) {
	p := DefaultWorldPolicyProfile()
	p.ResponseTimeout = -1
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for negative ResponseTimeout, got nil")
	}
}

func TestWorldPolicyProfile_Validate_MissingFingerprint(t *testing.T) {
	p := DefaultWorldPolicyProfile()
	p.PolicyFingerprint = ""
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing PolicyFingerprint, got nil")
	}
}

func TestComputePolicyFingerprint_Deterministic(t *testing.T) {
	p := DefaultWorldPolicyProfile()
	fp1 := ComputePolicyFingerprint(p)
	fp2 := ComputePolicyFingerprint(p)
	if fp1 != fp2 {
		t.Errorf("expected deterministic fingerprint, got %s and %s", fp1, fp2)
	}
	if fp1 == "" {
		t.Error("expected non-empty fingerprint")
	}
}

func TestComputePolicyFingerprint_ChangesWithPolicy(t *testing.T) {
	p1 := DefaultWorldPolicyProfile()
	p2 := DefaultWorldPolicyProfile()
	p2.MaximumInputLength = 100
	fp1 := ComputePolicyFingerprint(p1)
	fp2 := ComputePolicyFingerprint(p2)
	if fp1 == fp2 {
		t.Error("expected different fingerprints for different policies")
	}
}

// ============================================================================
// WorldCapabilities Tests
// ============================================================================

func TestDefaultWorldCapabilities_Valid(t *testing.T) {
	c := DefaultWorldCapabilities()
	if err := c.Validate(); err != nil {
		t.Fatalf("DefaultWorldCapabilities.Validate failed: %v", err)
	}
	if !c.SupportsText {
		t.Error("expected SupportsText=true for default capabilities")
	}
	if c.CapabilityFingerprint == "" {
		t.Error("expected non-empty CapabilityFingerprint")
	}
}

func TestWorldCapabilities_Validate_MissingFingerprint(t *testing.T) {
	c := DefaultWorldCapabilities()
	c.CapabilityFingerprint = ""
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing CapabilityFingerprint, got nil")
	}
}

func TestComputeCapabilityFingerprint_Deterministic(t *testing.T) {
	c := DefaultWorldCapabilities()
	fp1 := ComputeCapabilityFingerprint(c)
	fp2 := ComputeCapabilityFingerprint(c)
	if fp1 != fp2 {
		t.Errorf("expected deterministic capability fingerprint, got %s and %s", fp1, fp2)
	}
}

// ============================================================================
// Interaction Validation Tests
// ============================================================================

func buildValidInteraction() *Interaction {
	return &Interaction{
		InteractionID:   "test-interaction-001",
		SessionID:       "test-session-001",
		Origin:          OriginUser,
		Modality:        ModalityText,
		OriginalInput:   "Hello, IDUN!",
		NormalizedInput: "Hello, IDUN!",
		PayloadRef:      "sha256:abc123",
		CreatedAt:       time.Now().UTC(),
		ReplayMetadata: WorldReplayMetadata{
			WorldVersion:           WorldVersion,
			PolicyFingerprint:      "policy-fp-001",
			CapabilityFingerprint:  "cap-fp-001",
			InteractionFingerprint: "int-fp-001",
		},
	}
}

func TestInteraction_Validate_Valid(t *testing.T) {
	i := buildValidInteraction()
	if err := i.Validate(); err != nil {
		t.Fatalf("Validate failed on valid Interaction: %v", err)
	}
}

func TestInteraction_Validate_Nil(t *testing.T) {
	var i *Interaction
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for nil Interaction, got nil")
	}
}

func TestInteraction_Validate_MissingInteractionID(t *testing.T) {
	i := buildValidInteraction()
	i.InteractionID = ""
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for missing InteractionID, got nil")
	}
}

func TestInteraction_Validate_MissingSessionID(t *testing.T) {
	i := buildValidInteraction()
	i.SessionID = ""
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for missing SessionID, got nil")
	}
}

func TestInteraction_Validate_InvalidOrigin(t *testing.T) {
	i := buildValidInteraction()
	i.Origin = "UNKNOWN"
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for invalid Origin, got nil")
	}
}

func TestInteraction_Validate_InvalidModality(t *testing.T) {
	i := buildValidInteraction()
	i.Modality = "HOLOGRAPHIC"
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for invalid Modality, got nil")
	}
}

func TestInteraction_Validate_MissingOriginalInput(t *testing.T) {
	i := buildValidInteraction()
	i.OriginalInput = ""
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for missing OriginalInput, got nil")
	}
}

func TestInteraction_Validate_MissingPayloadRef(t *testing.T) {
	i := buildValidInteraction()
	i.PayloadRef = ""
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for missing PayloadRef, got nil")
	}
}

func TestInteraction_Validate_MissingReplayMetadata(t *testing.T) {
	i := buildValidInteraction()
	i.ReplayMetadata.PolicyFingerprint = ""
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for missing PolicyFingerprint in ReplayMetadata, got nil")
	}
}

// ============================================================================
// Response Validation Tests
// ============================================================================

func buildValidResponse() *Response {
	return &Response{
		ResponseID:    "test-response-001",
		InteractionID: "test-interaction-001",
		SessionID:     "test-session-001",
		Modality:      ModalityText,
		Content:       "Hello, human!",
		PayloadRef:    "sha256:response-payload",
		Status:        ResponseStatusSuccess,
		ResultStatus:  ResultStatusSuccess,
		CreatedAt:     time.Now().UTC(),
		ReplayMetadata: WorldReplayMetadata{
			WorldVersion:           WorldVersion,
			PolicyFingerprint:      "policy-fp-001",
			CapabilityFingerprint:  "cap-fp-001",
			InteractionFingerprint: "int-fp-001",
		},
	}
}

func TestResponse_Validate_Valid(t *testing.T) {
	r := buildValidResponse()
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate failed on valid Response: %v", err)
	}
}

func TestResponse_Validate_Nil(t *testing.T) {
	var r *Response
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for nil Response, got nil")
	}
}

func TestResponse_Validate_MissingResponseID(t *testing.T) {
	r := buildValidResponse()
	r.ResponseID = ""
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for missing ResponseID, got nil")
	}
}

func TestResponse_Validate_InvalidStatus(t *testing.T) {
	r := buildValidResponse()
	r.Status = "UNKNOWN"
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for invalid ResponseStatus, got nil")
	}
}

func TestResponse_Validate_InvalidResultStatus(t *testing.T) {
	r := buildValidResponse()
	r.ResultStatus = "UNKNOWN"
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for invalid ResultStatus, got nil")
	}
}

// ============================================================================
// InteractionFingerprint Tests (Refinement 9)
// ============================================================================

func TestComputeInteractionFingerprint_Deterministic(t *testing.T) {
	fp1 := ComputeInteractionFingerprint("hello", "hello", ModalityText, "policy-fp")
	fp2 := ComputeInteractionFingerprint("hello", "hello", ModalityText, "policy-fp")
	if fp1 != fp2 {
		t.Errorf("expected deterministic fingerprint, got %s and %s", fp1, fp2)
	}
}

func TestComputeInteractionFingerprint_ChangesWithInput(t *testing.T) {
	fp1 := ComputeInteractionFingerprint("hello", "hello", ModalityText, "policy-fp")
	fp2 := ComputeInteractionFingerprint("world", "world", ModalityText, "policy-fp")
	if fp1 == fp2 {
		t.Error("expected different fingerprints for different inputs")
	}
}

func TestComputeInteractionFingerprint_ChangesWithPolicy(t *testing.T) {
	fp1 := ComputeInteractionFingerprint("hello", "hello", ModalityText, "policy-fp-A")
	fp2 := ComputeInteractionFingerprint("hello", "hello", ModalityText, "policy-fp-B")
	if fp1 == fp2 {
		t.Error("expected different fingerprints for different policy fingerprints")
	}
}

// ============================================================================
// InteractionBuilder Tests
// ============================================================================

func TestInteractionBuilder_Build_Valid(t *testing.T) {
	meta := WorldReplayMetadata{
		WorldVersion:           WorldVersion,
		PolicyFingerprint:      "policy-fp",
		CapabilityFingerprint:  "cap-fp",
		InteractionFingerprint: "int-fp",
	}
	i, err := NewInteractionBuilder().
		WithSessionID("session-001").
		WithOrigin(OriginUser).
		WithModality(ModalityText).
		WithOriginalInput("Hello").
		WithNormalizedInput("Hello").
		WithPayloadRef("sha256:hello").
		WithReplayMetadata(meta).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if i.InteractionID == "" {
		t.Error("expected auto-generated InteractionID")
	}
	if i.SessionID != "session-001" {
		t.Errorf("expected SessionID=session-001, got %s", i.SessionID)
	}
}

func TestInteractionBuilder_Build_MissingRequired(t *testing.T) {
	// Missing SessionID
	_, err := NewInteractionBuilder().
		WithOrigin(OriginUser).
		WithModality(ModalityText).
		WithOriginalInput("Hello").
		WithPayloadRef("ref").
		Build()
	if err == nil {
		t.Fatal("expected error for missing SessionID, got nil")
	}
}

// ============================================================================
// ResponseBuilder Tests
// ============================================================================

func TestResponseBuilder_Build_Valid(t *testing.T) {
	meta := WorldReplayMetadata{
		WorldVersion:           WorldVersion,
		PolicyFingerprint:      "policy-fp",
		CapabilityFingerprint:  "cap-fp",
		InteractionFingerprint: "int-fp",
	}
	r, err := NewResponseBuilder().
		WithInteractionID("interaction-001").
		WithSessionID("session-001").
		WithModality(ModalityText).
		WithContent("Hello!").
		WithPayloadRef("sha256:hello").
		WithStatus(ResponseStatusSuccess).
		WithResultStatus(ResultStatusSuccess).
		WithReplayMetadata(meta).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if r.ResponseID == "" {
		t.Error("expected auto-generated ResponseID")
	}
}

func TestResponseBuilder_Build_MissingRequired(t *testing.T) {
	// Missing InteractionID
	_, err := NewResponseBuilder().
		WithSessionID("session-001").
		WithModality(ModalityText).
		WithStatus(ResponseStatusSuccess).
		WithResultStatus(ResultStatusSuccess).
		Build()
	if err == nil {
		t.Fatal("expected error for missing InteractionID, got nil")
	}
}

// ============================================================================
// WorldSummary Tests
// ============================================================================

func TestWorldSummary_Validate_Valid(t *testing.T) {
	s := &WorldSummary{
		TotalInteractions:   10,
		SuccessfulResponses: 9,
		FailedResponses:     1,
		AverageLatency:      100 * time.Millisecond,
		AverageInputLength:  42.5,
		AverageOutputLength: 120.0,
		TimeoutCount:        0,
		DroppedInputCount:   0,
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate failed on valid WorldSummary: %v", err)
	}
}

func TestWorldSummary_Validate_NegativeCount(t *testing.T) {
	s := &WorldSummary{TotalInteractions: -1}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for negative TotalInteractions, got nil")
	}
}

// ============================================================================
// Validation Helpers Tests
// ============================================================================

func TestNormalizeInput_TrimsAndCollapses(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  world  ", "hello world"},
		{"hello", "hello"},
		{"  ", ""},
		{"", ""},
		{"a\t\nb\r\nc", "a b c"},
		{"  leading", "leading"},
		{"trailing  ", "trailing"},
	}
	for _, tc := range tests {
		got := normalizeInput(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeInput(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestDropInvalidUTF8_PreservesValid(t *testing.T) {
	valid := "Hello, 世界! 🌍"
	result := dropInvalidUTF8(valid)
	if result != valid {
		t.Errorf("dropInvalidUTF8 modified valid UTF-8: got %q", result)
	}
}

func TestApplyPolicy_NormalizesAndEnforces(t *testing.T) {
	policy := DefaultWorldPolicyProfile()
	policy.MaximumInputLength = 20

	result, err := ApplyPolicy("  hello world  ", policy)
	if err != nil {
		t.Fatalf("ApplyPolicy failed: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", result)
	}

	// Test too-long input (after normalization)
	_, err = ApplyPolicy("this is a very long input that exceeds the limit", policy)
	if err == nil {
		t.Fatal("expected error for oversized input, got nil")
	}
}

func TestApplyPolicy_EmptyInputRejected(t *testing.T) {
	policy := DefaultWorldPolicyProfile()
	policy.AllowEmptyInput = false
	_, err := ApplyPolicy("   ", policy)
	if err == nil {
		t.Fatal("expected error for empty input when AllowEmptyInput=false, got nil")
	}
}

func TestApplyPolicy_EmptyInputAllowed(t *testing.T) {
	policy := DefaultWorldPolicyProfile()
	policy.AllowEmptyInput = true
	result, err := ApplyPolicy("   ", policy)
	if err != nil {
		t.Fatalf("unexpected error when AllowEmptyInput=true: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

// ============================================================================
// Enum Validity Tests
// ============================================================================

func TestModality_IsValid(t *testing.T) {
	valid := []Modality{ModalityText, ModalityVoice, ModalityVision, ModalityAPI}
	for _, m := range valid {
		if !m.IsValid() {
			t.Errorf("expected IsValid=true for %s", m)
		}
	}
	if Modality("UNKNOWN").IsValid() {
		t.Error("expected IsValid=false for unknown modality")
	}
}

func TestInteractionOrigin_IsValid(t *testing.T) {
	valid := []InteractionOrigin{OriginUser, OriginVoice, OriginVision, OriginAPI, OriginScheduler, OriginRobot, OriginSimulation}
	for _, o := range valid {
		if !o.IsValid() {
			t.Errorf("expected IsValid=true for %s", o)
		}
	}
	if InteractionOrigin("UNKNOWN").IsValid() {
		t.Error("expected IsValid=false for unknown origin")
	}
}

func TestInteractionResultStatus_IsValid(t *testing.T) {
	valid := []InteractionResultStatus{ResultStatusSuccess, ResultStatusFailed, ResultStatusTimeout, ResultStatusDropped, ResultStatusInterrupted}
	for _, r := range valid {
		if !r.IsValid() {
			t.Errorf("expected IsValid=true for %s", r)
		}
	}
}

func TestInteractionTerminationReason_IsValid(t *testing.T) {
	valid := []InteractionTerminationReason{
		TerminationInvalidInput, TerminationEmptyInput, TerminationWorkspaceFailure,
		TerminationExecutiveTimeout, TerminationUserCancelled, TerminationSystemShutdown,
	}
	for _, r := range valid {
		if !r.IsValid() {
			t.Errorf("expected IsValid=true for %s", r)
		}
	}
}

// ============================================================================
// WorldPolicyProfileBuilder Tests
// ============================================================================

func TestWorldPolicyProfileBuilder_Build_Valid(t *testing.T) {
	p, err := NewWorldPolicyProfileBuilder().
		WithMaximumInputLength(2048).
		WithAllowEmptyInput(false).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if p.MaximumInputLength != 2048 {
		t.Errorf("expected MaximumInputLength=2048, got %d", p.MaximumInputLength)
	}
	if p.PolicyFingerprint == "" {
		t.Error("expected non-empty PolicyFingerprint after Build")
	}
}
