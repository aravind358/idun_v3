package world

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// SchemaVersion defines the canonical invariant schema version for all World artifacts.
const SchemaVersion = "2.0.0-FROZEN"

// WorldVersion identifies the implementation version of the World subsystem.
const WorldVersion = "2.0.0-FROZEN"

// Bounded string length limits to prevent unbounded memory growth over decades.
const (
	MaxInputLength    = 65536 // 64 KiB maximum raw input
	MaxContentLength  = 65536 // 64 KiB maximum response content
	MaxSessionIDLen   = 256
	MaxAdapterNameLen = 128
)

// Sentinel errors for World domain artifacts and validation firewalls.
var (
	ErrNilInteraction        = errors.New("world: Interaction cannot be nil")
	ErrNilResponse           = errors.New("world: Response cannot be nil")
	ErrNilProfile            = errors.New("world: WorldPolicyProfile cannot be nil")
	ErrNilCapabilities       = errors.New("world: WorldCapabilities cannot be nil")
	ErrNilTrace              = errors.New("world: WorldTrace cannot be nil")
	ErrNilSummary            = errors.New("world: WorldSummary cannot be nil")
	ErrMissingInteractionID  = errors.New("world: InteractionID is required")
	ErrMissingSessionID      = errors.New("world: SessionID is required")
	ErrMissingOriginalInput  = errors.New("world: OriginalInput is required")
	ErrMissingPayloadRef     = errors.New("world: PayloadRef is required")
	ErrMissingResponseID     = errors.New("world: ResponseID is required")
	ErrMissingFingerprint    = errors.New("world: PolicyFingerprint or CapabilityFingerprint is required")
	ErrMissingPolicyVersion  = errors.New("world: PolicyVersion is required")
	ErrInvalidModality       = errors.New("world: invalid Modality")
	ErrInvalidOrigin         = errors.New("world: invalid InteractionOrigin")
	ErrInvalidResultStatus   = errors.New("world: invalid InteractionResultStatus")
	ErrInvalidResponseStatus = errors.New("world: invalid ResponseStatus")
	ErrInputTooLong          = errors.New("world: input exceeds maximum length")
	ErrContentTooLong        = errors.New("world: content exceeds maximum length")
	ErrServiceClosed         = errors.New("world: service is closed")
	ErrServiceNotStarted     = errors.New("world: service not started")
	ErrNilWorkspace          = errors.New("world: Global Workspace cannot be nil")
	ErrNilInputAdapter       = errors.New("world: InputAdapter cannot be nil")
	ErrNilOutputAdapter      = errors.New("world: OutputAdapter cannot be nil")
	ErrNilPayloadStorer      = errors.New("world: PayloadStorer cannot be nil")
)

// ============================================================================
// Modality
// ============================================================================

// Modality identifies the input/output channel type for an interaction.
// Modality describes the communication medium; InteractionOrigin describes the source.
type Modality string

const (
	// ModalityText represents plain-text terminal or keyboard input/output.
	ModalityText Modality = "TEXT"
	// ModalityVoice represents speech-to-text or text-to-speech channels.
	ModalityVoice Modality = "VOICE"
	// ModalityVision represents image or video input channels.
	ModalityVision Modality = "VISION"
	// ModalityAPI represents machine-to-machine structured API interactions.
	ModalityAPI Modality = "API"
)

// IsValid returns true if m is a recognized Modality.
func (m Modality) IsValid() bool {
	switch m {
	case ModalityText, ModalityVoice, ModalityVision, ModalityAPI:
		return true
	default:
		return false
	}
}

// ============================================================================
// InteractionOrigin
// ============================================================================

// InteractionOrigin identifies the source that originated an Interaction.
// This is independent from Modality. Understanding interprets content;
// World records only the origin without semantic evaluation.
type InteractionOrigin string

const (
	// OriginUser represents a human user at a terminal or UI.
	OriginUser InteractionOrigin = "USER"
	// OriginVoice represents a voice-enabled device or speech interface.
	OriginVoice InteractionOrigin = "VOICE"
	// OriginVision represents a camera or image-capture device.
	OriginVision InteractionOrigin = "VISION"
	// OriginAPI represents a machine client or external service.
	OriginAPI InteractionOrigin = "API"
	// OriginScheduler represents internally scheduled interactions.
	OriginScheduler InteractionOrigin = "SCHEDULER"
	// OriginRobot represents a robotics or actuator-driven input.
	OriginRobot InteractionOrigin = "ROBOT"
	// OriginSimulation represents synthetic interactions from test or simulation environments.
	OriginSimulation InteractionOrigin = "SIMULATION"
)

// IsValid returns true if o is a recognized InteractionOrigin.
func (o InteractionOrigin) IsValid() bool {
	switch o {
	case OriginUser, OriginVoice, OriginVision, OriginAPI, OriginScheduler, OriginRobot, OriginSimulation:
		return true
	default:
		return false
	}
}

// ============================================================================
// InteractionResultStatus and InteractionTerminationReason
// ============================================================================

// InteractionResultStatus is the high-level outcome of processing an Interaction.
// This is separate from InteractionTerminationReason, which explains factual cause.
type InteractionResultStatus string

const (
	// ResultStatusSuccess indicates the interaction was processed and a response was delivered.
	ResultStatusSuccess InteractionResultStatus = "SUCCESS"
	// ResultStatusFailed indicates the interaction failed to produce a response.
	ResultStatusFailed InteractionResultStatus = "FAILED"
	// ResultStatusTimeout indicates no response was received within the configured timeout.
	ResultStatusTimeout InteractionResultStatus = "TIMEOUT"
	// ResultStatusDropped indicates the interaction was dropped before dispatch (e.g., invalid input).
	ResultStatusDropped InteractionResultStatus = "DROPPED"
	// ResultStatusInterrupted indicates processing was interrupted by a higher-priority event.
	ResultStatusInterrupted InteractionResultStatus = "INTERRUPTED"
)

// IsValid returns true if r is a recognized InteractionResultStatus.
func (r InteractionResultStatus) IsValid() bool {
	switch r {
	case ResultStatusSuccess, ResultStatusFailed, ResultStatusTimeout, ResultStatusDropped, ResultStatusInterrupted:
		return true
	default:
		return false
	}
}

// InteractionTerminationReason explains the factual cause of a non-success interaction outcome.
// World records the reason; Reflection analyzes patterns across reasons.
type InteractionTerminationReason string

const (
	// TerminationInvalidInput indicates the interaction was rejected due to validation failure.
	TerminationInvalidInput InteractionTerminationReason = "INVALID_INPUT"
	// TerminationEmptyInput indicates the interaction was rejected due to empty or blank input.
	TerminationEmptyInput InteractionTerminationReason = "EMPTY_INPUT"
	// TerminationWorkspaceFailure indicates the Global Workspace rejected or failed to publish the envelope.
	TerminationWorkspaceFailure InteractionTerminationReason = "WORKSPACE_FAILURE"
	// TerminationExecutiveTimeout indicates Executive did not respond within the configured timeout.
	TerminationExecutiveTimeout InteractionTerminationReason = "EXECUTIVE_TIMEOUT"
	// TerminationUserCancelled indicates the user explicitly cancelled the interaction.
	TerminationUserCancelled InteractionTerminationReason = "USER_CANCELLED"
	// TerminationSystemShutdown indicates the service was shut down during interaction processing.
	TerminationSystemShutdown InteractionTerminationReason = "SYSTEM_SHUTDOWN"
)

// IsValid returns true if t is a recognized InteractionTerminationReason.
func (t InteractionTerminationReason) IsValid() bool {
	switch t {
	case TerminationInvalidInput, TerminationEmptyInput, TerminationWorkspaceFailure,
		TerminationExecutiveTimeout, TerminationUserCancelled, TerminationSystemShutdown:
		return true
	default:
		return false
	}
}

// ResponseStatus is the high-level status of an individual Response artifact.
type ResponseStatus string

const (
	ResponseStatusSuccess ResponseStatus = "SUCCESS"
	ResponseStatusError   ResponseStatus = "ERROR"
	ResponseStatusTimeout ResponseStatus = "TIMEOUT"
)

// IsValid returns true if s is a recognized ResponseStatus.
func (s ResponseStatus) IsValid() bool {
	switch s {
	case ResponseStatusSuccess, ResponseStatusError, ResponseStatusTimeout:
		return true
	default:
		return false
	}
}

// ============================================================================
// WorldPolicyProfile
// ============================================================================

// WorldPolicyProfile defines operational boundaries consumed by the World subsystem.
// This artifact is owned by Executive/Runtime and consumed read-only by World.
// World must never modify or derive policy; it only enforces what it receives.
type WorldPolicyProfile struct {
	PolicyVersion         string        `json:"policy_version"`
	SchemaVersion         string        `json:"schema_version"`
	MaximumInputLength    int           `json:"maximum_input_length"`
	MaximumResponseLength int           `json:"maximum_response_length"`
	NormalizeWhitespace   bool          `json:"normalize_whitespace"`
	DropInvalidUTF8       bool          `json:"drop_invalid_utf8"`
	AllowEmptyInput       bool          `json:"allow_empty_input"`
	ResponseTimeout       time.Duration `json:"response_timeout"`
	PolicyFingerprint     string        `json:"policy_fingerprint"`
}

// Validate verifies that the WorldPolicyProfile satisfies all structural invariants.
func (p *WorldPolicyProfile) Validate() error {
	if p == nil {
		return ErrNilProfile
	}
	if p.PolicyVersion == "" {
		return ErrMissingPolicyVersion
	}
	if p.SchemaVersion == "" {
		return fmt.Errorf("world: WorldPolicyProfile SchemaVersion is required")
	}
	if p.MaximumInputLength <= 0 {
		return fmt.Errorf("world: WorldPolicyProfile MaximumInputLength must be positive")
	}
	if p.MaximumResponseLength <= 0 {
		return fmt.Errorf("world: WorldPolicyProfile MaximumResponseLength must be positive")
	}
	if p.MaximumInputLength > MaxInputLength {
		return fmt.Errorf("world: WorldPolicyProfile MaximumInputLength %d exceeds absolute limit %d", p.MaximumInputLength, MaxInputLength)
	}
	if p.MaximumResponseLength > MaxContentLength {
		return fmt.Errorf("world: WorldPolicyProfile MaximumResponseLength %d exceeds absolute limit %d", p.MaximumResponseLength, MaxContentLength)
	}
	if p.ResponseTimeout < 0 {
		return fmt.Errorf("world: WorldPolicyProfile ResponseTimeout cannot be negative")
	}
	if p.PolicyFingerprint == "" {
		return ErrMissingFingerprint
	}
	return nil
}

// DefaultWorldPolicyProfile returns the standard, production-safe WorldPolicyProfile.
func DefaultWorldPolicyProfile() *WorldPolicyProfile {
	p := &WorldPolicyProfile{
		PolicyVersion:         "2.0.0-FROZEN",
		SchemaVersion:         SchemaVersion,
		MaximumInputLength:    4096,
		MaximumResponseLength: 8192,
		NormalizeWhitespace:   true,
		DropInvalidUTF8:       true,
		AllowEmptyInput:       false,
		ResponseTimeout:       30 * time.Second,
	}
	p.PolicyFingerprint = ComputePolicyFingerprint(p)
	return p
}

// ComputePolicyFingerprint derives a deterministic SHA-256 digest over structural policy fields.
func ComputePolicyFingerprint(p *WorldPolicyProfile) string {
	if p == nil {
		return ""
	}
	raw := fmt.Sprintf("world-policy|version:%s|schema:%s|maxInput:%d|maxResponse:%d|normalize:%v|dropUTF8:%v|allowEmpty:%v|timeout:%s",
		p.PolicyVersion,
		p.SchemaVersion,
		p.MaximumInputLength,
		p.MaximumResponseLength,
		p.NormalizeWhitespace,
		p.DropInvalidUTF8,
		p.AllowEmptyInput,
		p.ResponseTimeout.String(),
	)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// ============================================================================
// WorldCapabilities
// ============================================================================

// WorldCapabilities advertises deployment-specific I/O capabilities.
// Capabilities are immutable after service startup and describe what this
// deployment can receive and present; they do not affect cognitive processing.
type WorldCapabilities struct {
	SupportsText           bool   `json:"supports_text"`
	SupportsVoice          bool   `json:"supports_voice"`
	SupportsVision         bool   `json:"supports_vision"`
	SupportsAPI            bool   `json:"supports_api"`
	SupportsStreaming       bool   `json:"supports_streaming"`
	SupportsAttachments    bool   `json:"supports_attachments"`
	CapabilityFingerprint  string `json:"capability_fingerprint"`
}

// Validate verifies that the WorldCapabilities struct is well-formed.
func (c *WorldCapabilities) Validate() error {
	if c == nil {
		return ErrNilCapabilities
	}
	if c.CapabilityFingerprint == "" {
		return ErrMissingFingerprint
	}
	return nil
}

// DefaultWorldCapabilities returns the minimal capability set for a text-only deployment.
func DefaultWorldCapabilities() *WorldCapabilities {
	c := &WorldCapabilities{
		SupportsText:        true,
		SupportsVoice:       false,
		SupportsVision:      false,
		SupportsAPI:         false,
		SupportsStreaming:    false,
		SupportsAttachments: false,
	}
	c.CapabilityFingerprint = ComputeCapabilityFingerprint(c)
	return c
}

// ComputeCapabilityFingerprint derives a deterministic SHA-256 digest over capability flags.
func ComputeCapabilityFingerprint(c *WorldCapabilities) string {
	if c == nil {
		return ""
	}
	raw := fmt.Sprintf("world-caps|text:%v|voice:%v|vision:%v|api:%v|streaming:%v|attachments:%v",
		c.SupportsText,
		c.SupportsVoice,
		c.SupportsVision,
		c.SupportsAPI,
		c.SupportsStreaming,
		c.SupportsAttachments,
	)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// ============================================================================
// WorldReplayMetadata
// ============================================================================

// WorldReplayMetadata ensures deterministic replay across identical inputs and
// identical policy/capability configurations. Every Interaction and Response carries
// this metadata for traceability and audit.
type WorldReplayMetadata struct {
	WorldVersion           string `json:"world_version"`
	PolicyFingerprint      string `json:"policy_fingerprint"`
	CapabilityFingerprint  string `json:"capability_fingerprint"`
	InteractionFingerprint string `json:"interaction_fingerprint"`
	ReplaySeed             int64  `json:"replay_seed"`
}

// Validate verifies that the WorldReplayMetadata is well-formed.
func (r *WorldReplayMetadata) Validate() error {
	if r == nil {
		return fmt.Errorf("world: WorldReplayMetadata cannot be nil")
	}
	if r.WorldVersion == "" {
		return fmt.Errorf("world: WorldReplayMetadata.WorldVersion is required")
	}
	if r.PolicyFingerprint == "" || r.CapabilityFingerprint == "" {
		return ErrMissingFingerprint
	}
	return nil
}

// ============================================================================
// InteractionFingerprint
// ============================================================================

// ComputeInteractionFingerprint derives a deterministic SHA-256 digest over
// the canonical interaction content fields. Used for replay and diagnostics only;
// World never makes decisions based on this fingerprint.
func ComputeInteractionFingerprint(originalInput, normalizedInput string, modality Modality, policyFingerprint string) string {
	raw := fmt.Sprintf("interaction|original:%s|normalized:%s|modality:%s|policy:%s",
		originalInput,
		normalizedInput,
		string(modality),
		policyFingerprint,
	)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// ============================================================================
// Interaction
// ============================================================================

// Interaction is the canonical immutable domain artifact representing one user turn
// at the World boundary. It is created by the World service after normalization and
// validation of raw adapter input. Interaction is never modified after construction.
type Interaction struct {
	// InteractionID is the unique identifier for this interaction (content-addressed or UUID).
	InteractionID string `json:"interaction_id"`

	// SessionID groups multiple interactions into a logical conversation session.
	SessionID string `json:"session_id"`

	// Origin identifies the source of this interaction (User, API, Robot, etc.).
	Origin InteractionOrigin `json:"origin"`

	// Modality identifies the input channel type (Text, Voice, Vision, API).
	Modality Modality `json:"modality"`

	// OriginalInput stores exactly what the external adapter received, verbatim.
	OriginalInput string `json:"original_input"`

	// NormalizedInput stores the canonical form after whitespace normalization and UTF-8 validation.
	NormalizedInput string `json:"normalized_input"`

	// PayloadRef is the content-addressed storage URI for large input payloads.
	// For short inputs, this may be a hash of the NormalizedInput; for large inputs
	// it references persisted data in Core.Storage.
	PayloadRef string `json:"payload_ref"`

	// CreatedAt is the UTC timestamp of Interaction construction.
	CreatedAt time.Time `json:"created_at"`

	// ReplayMetadata provides deterministic replay provenance.
	ReplayMetadata WorldReplayMetadata `json:"replay_metadata"`
}

// Validate verifies that the Interaction satisfies all structural invariants.
func (i *Interaction) Validate() error {
	if i == nil {
		return ErrNilInteraction
	}
	if i.InteractionID == "" {
		return ErrMissingInteractionID
	}
	if i.SessionID == "" {
		return ErrMissingSessionID
	}
	if !i.Origin.IsValid() {
		return ErrInvalidOrigin
	}
	if !i.Modality.IsValid() {
		return ErrInvalidModality
	}
	if i.OriginalInput == "" {
		return ErrMissingOriginalInput
	}
	if len(i.OriginalInput) > MaxInputLength {
		return ErrInputTooLong
	}
	if i.PayloadRef == "" {
		return ErrMissingPayloadRef
	}
	return i.ReplayMetadata.Validate()
}

// ============================================================================
// Response
// ============================================================================

// Response is the canonical immutable domain artifact representing the system's
// reply to an Interaction. It is created by the World service upon receiving
// a response envelope from the Global Workspace and is presented via OutputAdapter.
type Response struct {
	// ResponseID is the unique identifier for this response.
	ResponseID string `json:"response_id"`

	// InteractionID links this response to its originating Interaction.
	InteractionID string `json:"interaction_id"`

	// SessionID groups responses within a logical conversation session.
	SessionID string `json:"session_id"`

	// Modality identifies the output channel type (Text, Voice, etc.).
	Modality Modality `json:"modality"`

	// Content contains the human-readable response text or structured payload.
	Content string `json:"content"`

	// PayloadRef is the content-addressed storage URI for large response payloads.
	PayloadRef string `json:"payload_ref"`

	// Status is the high-level outcome for this individual response.
	Status ResponseStatus `json:"status"`

	// ResultStatus is the outcome of the entire interaction lifecycle.
	ResultStatus InteractionResultStatus `json:"result_status"`

	// TerminationReason explains why a non-success result occurred (empty on success).
	TerminationReason InteractionTerminationReason `json:"termination_reason,omitempty"`

	// CreatedAt is the UTC timestamp of Response construction.
	CreatedAt time.Time `json:"created_at"`

	// ExecutionDuration records the wall-clock time from Interaction creation to Response delivery.
	ExecutionDuration time.Duration `json:"execution_duration"`

	// ReplayMetadata provides deterministic replay provenance.
	ReplayMetadata WorldReplayMetadata `json:"replay_metadata"`
}

// Validate verifies that the Response satisfies all structural invariants.
func (r *Response) Validate() error {
	if r == nil {
		return ErrNilResponse
	}
	if r.ResponseID == "" {
		return ErrMissingResponseID
	}
	if r.InteractionID == "" {
		return ErrMissingInteractionID
	}
	if r.SessionID == "" {
		return ErrMissingSessionID
	}
	if !r.Modality.IsValid() {
		return ErrInvalidModality
	}
	if !r.Status.IsValid() {
		return ErrInvalidResponseStatus
	}
	if !r.ResultStatus.IsValid() {
		return ErrInvalidResultStatus
	}
	if len(r.Content) > MaxContentLength {
		return ErrContentTooLong
	}
	return r.ReplayMetadata.Validate()
}

// ============================================================================
// WorldTrace
// ============================================================================

// WorldTrace is the immutable diagnostic telemetry artifact produced for each interaction.
// It is write-only from World's perspective — World records it; Reflection analyzes it.
// World must never evaluate or reason about its own traces.
type WorldTrace struct {
	// InteractionID links the trace to its originating Interaction.
	InteractionID string `json:"interaction_id"`

	// InteractionFingerprint is the deterministic SHA-256 digest over interaction content.
	InteractionFingerprint string `json:"interaction_fingerprint"`

	// AdapterName identifies the InputAdapter that received this interaction.
	AdapterName string `json:"adapter_name"`

	// AdapterVersion identifies the version of the InputAdapter implementation.
	AdapterVersion string `json:"adapter_version"`

	// AdapterFingerprint is a deterministic SHA-256 identity digest for the adapter implementation.
	AdapterFingerprint string `json:"adapter_fingerprint"`

	// Origin records which external source produced the interaction.
	Origin InteractionOrigin `json:"origin"`

	// InputModality identifies the input channel type.
	InputModality Modality `json:"input_modality"`

	// OutputModality identifies the output channel type.
	OutputModality Modality `json:"output_modality"`

	// ExecutionTime records the wall-clock duration for processing this interaction.
	ExecutionTime time.Duration `json:"execution_time"`

	// WorldVersion records the implementation version of the World subsystem.
	WorldVersion string `json:"world_version"`

	// PolicyFingerprint records which policy governed this interaction.
	PolicyFingerprint string `json:"policy_fingerprint"`

	// CapabilityFingerprint records which capability set governed this interaction.
	CapabilityFingerprint string `json:"capability_fingerprint"`

	// ResultStatus is the final outcome status.
	ResultStatus InteractionResultStatus `json:"result_status"`

	// TerminationReason is the factual cause for non-success outcomes.
	TerminationReason InteractionTerminationReason `json:"termination_reason,omitempty"`

	// ReplayMetadata provides deterministic replay provenance.
	ReplayMetadata WorldReplayMetadata `json:"replay_metadata"`

	// Timestamp is when this trace was recorded.
	Timestamp time.Time `json:"timestamp"`
}

// Validate verifies that the WorldTrace satisfies all structural invariants.
func (t *WorldTrace) Validate() error {
	if t == nil {
		return ErrNilTrace
	}
	if t.InteractionID == "" {
		return ErrMissingInteractionID
	}
	if t.WorldVersion == "" {
		return fmt.Errorf("world: WorldTrace.WorldVersion is required")
	}
	if t.PolicyFingerprint == "" || t.CapabilityFingerprint == "" {
		return ErrMissingFingerprint
	}
	if !t.ResultStatus.IsValid() {
		return ErrInvalidResultStatus
	}
	return t.ReplayMetadata.Validate()
}

// ============================================================================
// WorldSummary
// ============================================================================

// WorldSummary provides bounded statistical bookkeeping across all interactions.
// This is pure recording — World counts facts; Reflection interprets patterns.
type WorldSummary struct {
	TotalInteractions   int64         `json:"total_interactions"`
	SuccessfulResponses int64         `json:"successful_responses"`
	FailedResponses     int64         `json:"failed_responses"`
	AverageLatency      time.Duration `json:"average_latency"`
	AverageInputLength  float64       `json:"average_input_length"`
	AverageOutputLength float64       `json:"average_output_length"`
	TimeoutCount        int64         `json:"timeout_count"`
	DroppedInputCount   int64         `json:"dropped_input_count"`
}

// Validate verifies that the WorldSummary is well-formed.
func (s *WorldSummary) Validate() error {
	if s == nil {
		return ErrNilSummary
	}
	if s.TotalInteractions < 0 || s.SuccessfulResponses < 0 || s.FailedResponses < 0 ||
		s.TimeoutCount < 0 || s.DroppedInputCount < 0 {
		return fmt.Errorf("world: WorldSummary counts cannot be negative")
	}
	if s.AverageInputLength < 0 || s.AverageOutputLength < 0 {
		return fmt.Errorf("world: WorldSummary average lengths cannot be negative")
	}
	return nil
}

// ============================================================================
// ID generation
// ============================================================================

// generateID returns a secure random 16-byte hex identifier.
func generateID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
