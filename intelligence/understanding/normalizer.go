package understanding

import (
	"strings"
	"unicode"
)

// NormalizedText holds the cleaned, tokenized representation of a perception utterance.
type NormalizedText struct {
	// Original records the raw input string.
	Original string

	// Cleaned records the trimmed, lowercase, whitespace-collapsed string.
	Cleaned string

	// Tokens records whitespace/punctuation-split lowercase words.
	Tokens []string

	// PronounsFound records deictic and pronominal anchors detected ("it", "her", "him", "that", "them").
	PronounsFound []string
}

// Normalizer defines the capability to deterministically normalize raw utterance text.
type Normalizer interface {
	Normalize(rawText string) NormalizedText
}

// DefaultNormalizer is the canonical deterministic NLP normalizer.
type DefaultNormalizer struct{}

// NewDefaultNormalizer constructs a new DefaultNormalizer.
func NewDefaultNormalizer() *DefaultNormalizer {
	return &DefaultNormalizer{}
}

// Normalize cleans text, tokenizes, and detects pronominal/deictic anchors.
func (n *DefaultNormalizer) Normalize(rawText string) NormalizedText {
	trimmed := strings.TrimSpace(rawText)
	var sb strings.Builder
	var lastWasSpace bool

	for _, r := range trimmed {
		if unicode.IsSpace(r) {
			if !lastWasSpace {
				sb.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			sb.WriteRune(unicode.ToLower(r))
			lastWasSpace = false
		}
	}

	cleaned := sb.String()
	rawTokens := strings.FieldsFunc(cleaned, func(r rune) bool {
		return unicode.IsSpace(r) || (unicode.IsPunct(r) && r != '_' && r != '-')
	})

	var pronouns []string
	pronounSet := map[string]struct{}{
		"it":   {},
		"her":  {},
		"him":  {},
		"them": {},
		"that": {},
		"this": {},
	}

	for _, tok := range rawTokens {
		if _, ok := pronounSet[tok]; ok {
			pronouns = append(pronouns, tok)
		}
	}

	return NormalizedText{
		Original:      rawText,
		Cleaned:       cleaned,
		Tokens:        rawTokens,
		PronounsFound: pronouns,
	}
}

// Ensure DefaultNormalizer implements Normalizer.
var _ Normalizer = (*DefaultNormalizer)(nil)
