package splitter

import (
	"regexp"
	"strings"
)

// Splitter defines the interface for splitting utterances into distinct goals.
type Splitter interface {
	// Split takes an utterance and a validation function (which checks if a string is a valid goal).
	// It returns a slice of distinct goal utterances.
	Split(utterance string, isValidGoal func(string) bool) []string
}

// deterministicSplitter implements Splitter using recognized connectors.
type deterministicSplitter struct {
	registry *ConnectorRegistry
}

// NewDeterministicSplitter creates a new splitter with approved connectors from the registry.
func NewDeterministicSplitter(registry *ConnectorRegistry) Splitter {
	if registry == nil {
		registry = NewConnectorRegistry()
	}
	return &deterministicSplitter{
		registry: registry,
	}
}

// Split attempts to split the utterance by connectors using a greedy left-to-right O(N) scan.
func (s *deterministicSplitter) Split(utterance string, isValidGoal func(string) bool) []string {
	// Build pattern dynamically from the registry
	connectors := s.registry.GetConnectors()
	escaped := make([]string, len(connectors))
	for i, c := range connectors {
		escaped[i] = regexp.QuoteMeta(c)
	}
	pattern := regexp.MustCompile(`(?i)(?:` + strings.Join(escaped, `|`) + `)`)

	matches := pattern.FindAllStringIndex(utterance, -1)
	if len(matches) == 0 {
		return []string{strings.TrimSpace(utterance)}
	}

	var results []string
	currentStart := 0
	
	for i := 0; i < len(matches); i++ {
		// Test the candidate phrase from currentStart to the start of this connector
		candidateEnd := matches[i][0]
		candidate := strings.TrimSpace(utterance[currentStart:candidateEnd])
		
		if candidate != "" && isValidGoal(candidate) {
			results = append(results, candidate)
			// Move start past this connector
			currentStart = matches[i][1]
		}
		// If it's NOT a valid goal, we do NOT split here.
		// We just leave currentStart where it is, effectively including this connector
		// and the current chunk in the ongoing phrase.
	}

	// Add the remaining part of the string
	finalChunk := strings.TrimSpace(utterance[currentStart:])
	if finalChunk != "" {
		// If we've found prior goals, we just append the remainder.
		// A strict implementation might validate the final chunk too, but if it fails,
		// there's no more utterance to append to. For robustness, if we split at least once,
		// we append the remainder as a goal (even if flawed, it's what's left).
		// If we never found a split, we return the whole utterance.
		if len(results) > 0 {
			results = append(results, finalChunk)
		} else {
			return []string{strings.TrimSpace(utterance)}
		}
	} else if len(results) == 0 {
		return []string{strings.TrimSpace(utterance)}
	}

	return results
}
