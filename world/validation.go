package world

import (
	"fmt"
	"time"
	"unicode/utf8"
)

// ============================================================================
// Input Normalization
// ============================================================================

// normalizeInput applies canonical whitespace normalization to raw adapter input.
// It trims leading/trailing whitespace and collapses internal runs of whitespace
// to a single space. Operates only on well-formed UTF-8 strings.
func normalizeInput(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	result := make([]rune, 0, len(runes))
	inSpace := false
	leadingDone := false

	for _, r := range runes {
		isSpace := r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f'
		if isSpace {
			if leadingDone && !inSpace {
				result = append(result, ' ')
				inSpace = true
			}
		} else {
			result = append(result, r)
			inSpace = false
			leadingDone = true
		}
	}

	// Trim trailing space added by collapse
	if len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}
	return string(result)
}

// dropInvalidUTF8 removes bytes that form invalid UTF-8 sequences.
// Valid UTF-8 code points are preserved; invalid bytes are silently dropped.
func dropInvalidUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	b := []byte(s)
	out := make([]rune, 0, len(b))
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r != utf8.RuneError || size != 1 {
			out = append(out, r)
		}
		i += size
	}
	return string(out)
}

// ApplyPolicy applies the WorldPolicyProfile normalization rules to raw adapter input,
// returning the canonical form. This is called by the World service during Interaction
// construction; adapters must not call it directly.
func ApplyPolicy(rawInput string, profile *WorldPolicyProfile) (string, error) {
	if profile == nil {
		return rawInput, nil
	}
	result := rawInput
	if profile.DropInvalidUTF8 {
		result = dropInvalidUTF8(result)
	}
	if profile.NormalizeWhitespace {
		result = normalizeInput(result)
	}
	if !profile.AllowEmptyInput && result == "" {
		return "", fmt.Errorf("%w: input is empty after normalization", ErrMissingOriginalInput)
	}
	if len(result) > profile.MaximumInputLength {
		return "", fmt.Errorf("%w: normalized input length %d exceeds policy maximum %d", ErrInputTooLong, len(result), profile.MaximumInputLength)
	}
	return result, nil
}

// validateModality checks that the provided Modality is recognized.
func validateModality(m Modality) error {
	if !m.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidModality, string(m))
	}
	return nil
}

// validateOrigin checks that the provided InteractionOrigin is recognized.
func validateOrigin(o InteractionOrigin) error {
	if !o.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidOrigin, string(o))
	}
	return nil
}

// validateStringBound returns an error if s is longer than maxLen bytes.
func validateStringBound(fieldName, s string, maxLen int) error {
	if len(s) > maxLen {
		return fmt.Errorf("world: field %s length %d exceeds maximum %d", fieldName, len(s), maxLen)
	}
	return nil
}

// buildWorldReplayMetadata constructs a WorldReplayMetadata for an Interaction.
func buildWorldReplayMetadata(interactionFingerprint, policyFingerprint, capabilityFingerprint string, seed int64) WorldReplayMetadata {
	return WorldReplayMetadata{
		WorldVersion:           WorldVersion,
		PolicyFingerprint:      policyFingerprint,
		CapabilityFingerprint:  capabilityFingerprint,
		InteractionFingerprint: interactionFingerprint,
		ReplaySeed:             seed,
	}
}

// buildWorldTrace constructs a WorldTrace from interaction context.
func buildWorldTrace(
	interaction *Interaction,
	adapter InputAdapter,
	outputModality Modality,
	execTime time.Duration,
	result InteractionResultStatus,
	termination InteractionTerminationReason,
) *WorldTrace {
	adapterName := ""
	adapterVersion := ""
	adapterFP := ""
	if adapter != nil {
		adapterName = adapter.Name()
		adapterVersion = adapter.AdapterVersion()
		adapterFP = adapter.AdapterFingerprint()
	}
	return &WorldTrace{
		InteractionID:          interaction.InteractionID,
		InteractionFingerprint: interaction.ReplayMetadata.InteractionFingerprint,
		AdapterName:            adapterName,
		AdapterVersion:         adapterVersion,
		AdapterFingerprint:     adapterFP,
		Origin:                 interaction.Origin,
		InputModality:          interaction.Modality,
		OutputModality:         outputModality,
		ExecutionTime:          execTime,
		WorldVersion:           WorldVersion,
		PolicyFingerprint:      interaction.ReplayMetadata.PolicyFingerprint,
		CapabilityFingerprint:  interaction.ReplayMetadata.CapabilityFingerprint,
		ResultStatus:           result,
		TerminationReason:      termination,
		ReplayMetadata:         interaction.ReplayMetadata,
		Timestamp:              time.Now().UTC(),
	}
}
