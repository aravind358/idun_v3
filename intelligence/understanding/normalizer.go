package understanding

import (
	"regexp"
	"strings"
	"unicode"
)

// NormalizationProfile tracks the deterministic corrections and filler words removed
// during normalization, ensuring the original utterance remains recoverable.
type NormalizationProfile struct {
	TyposCorrected map[string]string // Original typo -> Corrected form
	FillersRemoved []string          // Conversational filler phrases removed
	SynonymsSubstituted map[string]string // Original phrase -> Canonical synonym
}

// NormalizedText holds the cleaned, tokenized representation of a perception utterance.
type NormalizedText struct {
	// Original records the raw input string.
	Original string

	// Cleaned records the trimmed, lowercase, whitespace-collapsed string,
	// with typos corrected and filler words stripped.
	Cleaned string

	// Tokens records whitespace/punctuation-split lowercase words from the Cleaned string.
	Tokens []string

	// PronounsFound records deictic and pronominal anchors detected ("it", "her", "him", "that", "them").
	PronounsFound []string

	// Profile records what transformations occurred during normalization.
	Profile NormalizationProfile
}

// Normalizer defines the capability to deterministically normalize raw utterance text.
type Normalizer interface {
	Normalize(rawText string) NormalizedText
}

// DefaultNormalizer is the canonical deterministic NLP normalizer.
type DefaultNormalizer struct {
	typos   map[string]string
	fillers []string
	synonyms map[string][]string
}

// NewDefaultNormalizer constructs a new DefaultNormalizer with the default registry.
func NewDefaultNormalizer() *DefaultNormalizer {
	return &DefaultNormalizer{
		typos:   DefaultTypos,
		fillers: DefaultFillers,
		synonyms: DefaultSynonyms,
	}
}

// NewNormalizer constructs a Normalizer with custom typo and filler dictionaries.
func NewNormalizer(typos map[string]string, fillers []string) *DefaultNormalizer {
	return &DefaultNormalizer{
		typos:   typos,
		fillers: fillers,
		synonyms: DefaultSynonyms,
	}
}

// Normalize cleans text, corrects typos, strips filler words, tokenizes, and detects pronominal anchors.
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
	profile := NormalizationProfile{
		TyposCorrected: make(map[string]string),
		FillersRemoved: make([]string, 0),
		SynonymsSubstituted: make(map[string]string),
	}

	// 1. Apply typo correction
	if len(n.typos) > 0 {
		tokens := strings.Fields(cleaned)
		for i, tok := range tokens {
			if correction, ok := n.typos[tok]; ok {
				profile.TyposCorrected[tok] = correction
				tokens[i] = correction
			}
		}
		cleaned = strings.Join(tokens, " ")
	}

	// 2. Apply filler stripping
	// Iterate through fillers and remove them using regex for word boundaries
	if len(n.fillers) > 0 {
		for _, f := range n.fillers {
			re := regexp.MustCompile(`\b` + f + `\b`)
			if re.MatchString(cleaned) {
				profile.FillersRemoved = append(profile.FillersRemoved, f)
				cleaned = re.ReplaceAllString(cleaned, "")
			}
		}
		// Clean up any extra spaces left behind
		cleaned = strings.Join(strings.Fields(cleaned), " ")
	}

	
	// 3. Apply synonym substitution
	if len(n.synonyms) > 0 {
		for canonical, synList := range n.synonyms {
			for _, syn := range synList {
				// use word boundary regex to avoid partial matches
				re := regexp.MustCompile(`\b` + syn + `\b`)
				if re.MatchString(cleaned) {
					profile.SynonymsSubstituted[syn] = canonical
					cleaned = re.ReplaceAllString(cleaned, canonical)
				}
			}
		}
		cleaned = strings.Join(strings.Fields(cleaned), " ")
	}

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
		Profile:       profile,
	}
}

// Ensure DefaultNormalizer implements Normalizer.
var _ Normalizer = (*DefaultNormalizer)(nil)
