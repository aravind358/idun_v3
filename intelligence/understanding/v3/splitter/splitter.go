package splitter

import (
	"regexp"
	"strings"
)

// Splitter defines the interface for splitting utterances into distinct goals.
type Splitter interface {
	// Split takes an utterance and a validation function (which checks if a string is a valid goal).
	// It returns a slice of distinct goal utterances. If the utterance cannot be split into
	// multiple valid goals, it returns the original utterance as a single-element slice.
	Split(utterance string, isValidGoal func(string) bool) []string
}

// deterministicSplitter implements Splitter using recognized connectors.
type deterministicSplitter struct {
	connectors []string
	pattern    *regexp.Regexp
}

// NewDeterministicSplitter creates a new splitter with approved connectors.
func NewDeterministicSplitter() Splitter {
	// Order matters: longer connectors first to prevent partial matches.
	connectors := []string{
		" and then ",
		" after that ",
		" as well as ",
		" and ",
		" then ",
		" also ",
		", and ",
		", then ",
		", ",
	}
	
	escaped := make([]string, len(connectors))
	for i, c := range connectors {
		escaped[i] = regexp.QuoteMeta(c)
	}
	
	pattern := regexp.MustCompile(`(?i)(?:` + strings.Join(escaped, `|`) + `)`)
	
	return &deterministicSplitter{
		connectors: connectors,
		pattern:    pattern,
	}
}

// Split attempts to split the utterance by connectors. It validates the splits.
func (s *deterministicSplitter) Split(utterance string, isValidGoal func(string) bool) []string {
	matches := s.pattern.FindAllStringIndex(utterance, -1)
	if len(matches) == 0 {
		return []string{strings.TrimSpace(utterance)}
	}
	
	n := len(matches)
	// Iterate through all 2^n possible split combinations.
	// We want the maximum number of valid splits.
	var bestSplit []string
	bestCount := 0

	// 0 is no splits, which we can check as fallback, but we start by checking combinations with splits
	for i := (1 << n) - 1; i > 0; i-- {
		chunks := s.getChunks(utterance, matches, i)
		
		allValid := true
		for _, chunk := range chunks {
			if !isValidGoal(chunk) {
				allValid = false
				break
			}
		}
		
		if allValid && len(chunks) > bestCount {
			bestSplit = chunks
			bestCount = len(chunks)
		}
	}
	
	if bestCount > 0 {
		return bestSplit
	}
	
	return []string{strings.TrimSpace(utterance)}
}

// getChunks generates strings by splitting the utterance at the connectors selected by the bitmask.
func (s *deterministicSplitter) getChunks(utterance string, matches [][]int, bitmask int) []string {
	var chunks []string
	lastEnd := 0
	
	for i := 0; i < len(matches); i++ {
		if (bitmask & (1 << (len(matches) - 1 - i))) != 0 {
			chunk := utterance[lastEnd:matches[i][0]]
			chunks = append(chunks, strings.TrimSpace(chunk))
			lastEnd = matches[i][1]
		}
	}
	
	chunk := utterance[lastEnd:]
	if chunk != "" {
		chunks = append(chunks, strings.TrimSpace(chunk))
	}
	
	return chunks
}
