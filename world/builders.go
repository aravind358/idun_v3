package world

import (
	"fmt"
	"time"
)

// ============================================================================
// InteractionBuilder
// ============================================================================

// InteractionBuilder is a fluent, validated builder for Interaction artifacts.
// Every .Build() call runs .Validate() and fails fast on invalid state.
type InteractionBuilder struct {
	interaction Interaction
}

// NewInteractionBuilder initializes a builder with a new random InteractionID and current UTC timestamp.
func NewInteractionBuilder() *InteractionBuilder {
	return &InteractionBuilder{
		interaction: Interaction{
			InteractionID: generateID(),
			CreatedAt:     time.Now().UTC(),
		},
	}
}

// WithInteractionID overrides the auto-generated InteractionID.
func (b *InteractionBuilder) WithInteractionID(id string) *InteractionBuilder {
	b.interaction.InteractionID = id
	return b
}

// WithSessionID sets the session grouping identifier.
func (b *InteractionBuilder) WithSessionID(sessionID string) *InteractionBuilder {
	b.interaction.SessionID = sessionID
	return b
}

// WithOrigin sets the InteractionOrigin.
func (b *InteractionBuilder) WithOrigin(origin InteractionOrigin) *InteractionBuilder {
	b.interaction.Origin = origin
	return b
}

// WithModality sets the Modality for this interaction.
func (b *InteractionBuilder) WithModality(modality Modality) *InteractionBuilder {
	b.interaction.Modality = modality
	return b
}

// WithOriginalInput sets OriginalInput to exactly what the adapter received.
func (b *InteractionBuilder) WithOriginalInput(input string) *InteractionBuilder {
	b.interaction.OriginalInput = input
	return b
}

// WithNormalizedInput sets NormalizedInput to the canonical form.
func (b *InteractionBuilder) WithNormalizedInput(input string) *InteractionBuilder {
	b.interaction.NormalizedInput = input
	return b
}

// WithPayloadRef sets the content-addressed storage URI for the interaction payload.
func (b *InteractionBuilder) WithPayloadRef(ref string) *InteractionBuilder {
	b.interaction.PayloadRef = ref
	return b
}

// WithReplayMetadata sets the WorldReplayMetadata for deterministic replay.
func (b *InteractionBuilder) WithReplayMetadata(meta WorldReplayMetadata) *InteractionBuilder {
	b.interaction.ReplayMetadata = meta
	return b
}

// WithCreatedAt overrides the auto-set creation timestamp.
func (b *InteractionBuilder) WithCreatedAt(t time.Time) *InteractionBuilder {
	b.interaction.CreatedAt = t
	return b
}

// Build validates and returns the immutable Interaction.
func (b *InteractionBuilder) Build() (*Interaction, error) {
	if err := b.interaction.Validate(); err != nil {
		return nil, fmt.Errorf("world: InteractionBuilder.Build: %w", err)
	}
	copy := b.interaction
	return &copy, nil
}

// ============================================================================
// ResponseBuilder
// ============================================================================

// ResponseBuilder is a fluent, validated builder for Response artifacts.
// Every .Build() call runs .Validate() and fails fast on invalid state.
type ResponseBuilder struct {
	response Response
}

// NewResponseBuilder initializes a builder with a new random ResponseID and current UTC timestamp.
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		response: Response{
			ResponseID: generateID(),
			CreatedAt:  time.Now().UTC(),
		},
	}
}

// WithResponseID overrides the auto-generated ResponseID.
func (b *ResponseBuilder) WithResponseID(id string) *ResponseBuilder {
	b.response.ResponseID = id
	return b
}

// WithInteractionID sets the InteractionID this response corresponds to.
func (b *ResponseBuilder) WithInteractionID(id string) *ResponseBuilder {
	b.response.InteractionID = id
	return b
}

// WithSessionID sets the session identifier.
func (b *ResponseBuilder) WithSessionID(sessionID string) *ResponseBuilder {
	b.response.SessionID = sessionID
	return b
}

// WithModality sets the output Modality.
func (b *ResponseBuilder) WithModality(modality Modality) *ResponseBuilder {
	b.response.Modality = modality
	return b
}

// WithContent sets the human-readable response content.
func (b *ResponseBuilder) WithContent(content string) *ResponseBuilder {
	b.response.Content = content
	return b
}

// WithPayloadRef sets the content-addressed storage URI for large response payloads.
func (b *ResponseBuilder) WithPayloadRef(ref string) *ResponseBuilder {
	b.response.PayloadRef = ref
	return b
}

// WithStatus sets the ResponseStatus.
func (b *ResponseBuilder) WithStatus(status ResponseStatus) *ResponseBuilder {
	b.response.Status = status
	return b
}

// WithResultStatus sets the InteractionResultStatus for the full interaction lifecycle.
func (b *ResponseBuilder) WithResultStatus(status InteractionResultStatus) *ResponseBuilder {
	b.response.ResultStatus = status
	return b
}

// WithTerminationReason sets the InteractionTerminationReason for non-success outcomes.
func (b *ResponseBuilder) WithTerminationReason(reason InteractionTerminationReason) *ResponseBuilder {
	b.response.TerminationReason = reason
	return b
}

// WithExecutionDuration sets the wall-clock duration from Interaction creation to Response delivery.
func (b *ResponseBuilder) WithExecutionDuration(d time.Duration) *ResponseBuilder {
	b.response.ExecutionDuration = d
	return b
}

// WithReplayMetadata sets the WorldReplayMetadata for deterministic replay.
func (b *ResponseBuilder) WithReplayMetadata(meta WorldReplayMetadata) *ResponseBuilder {
	b.response.ReplayMetadata = meta
	return b
}

// WithCreatedAt overrides the auto-set creation timestamp.
func (b *ResponseBuilder) WithCreatedAt(t time.Time) *ResponseBuilder {
	b.response.CreatedAt = t
	return b
}

// Build validates and returns the immutable Response.
func (b *ResponseBuilder) Build() (*Response, error) {
	if err := b.response.Validate(); err != nil {
		return nil, fmt.Errorf("world: ResponseBuilder.Build: %w", err)
	}
	copy := b.response
	return &copy, nil
}

// ============================================================================
// WorldPolicyProfileBuilder
// ============================================================================

// WorldPolicyProfileBuilder is a fluent, validated builder for WorldPolicyProfile.
type WorldPolicyProfileBuilder struct {
	profile WorldPolicyProfile
}

// NewWorldPolicyProfileBuilder initializes a builder with default profile values.
func NewWorldPolicyProfileBuilder() *WorldPolicyProfileBuilder {
	def := DefaultWorldPolicyProfile()
	return &WorldPolicyProfileBuilder{profile: *def}
}

// WithPolicyVersion sets the policy version string.
func (b *WorldPolicyProfileBuilder) WithPolicyVersion(version string) *WorldPolicyProfileBuilder {
	b.profile.PolicyVersion = version
	return b
}

// WithMaximumInputLength sets the maximum input length in bytes.
func (b *WorldPolicyProfileBuilder) WithMaximumInputLength(max int) *WorldPolicyProfileBuilder {
	b.profile.MaximumInputLength = max
	return b
}

// WithMaximumResponseLength sets the maximum response length in bytes.
func (b *WorldPolicyProfileBuilder) WithMaximumResponseLength(max int) *WorldPolicyProfileBuilder {
	b.profile.MaximumResponseLength = max
	return b
}

// WithNormalizeWhitespace configures whitespace normalization.
func (b *WorldPolicyProfileBuilder) WithNormalizeWhitespace(normalize bool) *WorldPolicyProfileBuilder {
	b.profile.NormalizeWhitespace = normalize
	return b
}

// WithDropInvalidUTF8 configures whether invalid UTF-8 bytes are silently dropped.
func (b *WorldPolicyProfileBuilder) WithDropInvalidUTF8(drop bool) *WorldPolicyProfileBuilder {
	b.profile.DropInvalidUTF8 = drop
	return b
}

// WithAllowEmptyInput configures whether empty inputs are accepted.
func (b *WorldPolicyProfileBuilder) WithAllowEmptyInput(allow bool) *WorldPolicyProfileBuilder {
	b.profile.AllowEmptyInput = allow
	return b
}

// WithResponseTimeout sets the response timeout duration.
func (b *WorldPolicyProfileBuilder) WithResponseTimeout(d time.Duration) *WorldPolicyProfileBuilder {
	b.profile.ResponseTimeout = d
	return b
}

// Build recomputes the PolicyFingerprint, validates, and returns the immutable profile.
func (b *WorldPolicyProfileBuilder) Build() (*WorldPolicyProfile, error) {
	b.profile.PolicyFingerprint = ComputePolicyFingerprint(&b.profile)
	if err := b.profile.Validate(); err != nil {
		return nil, fmt.Errorf("world: WorldPolicyProfileBuilder.Build: %w", err)
	}
	copy := b.profile
	return &copy, nil
}
