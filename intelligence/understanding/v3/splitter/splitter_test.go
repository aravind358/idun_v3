package splitter

import (
	"reflect"
	"testing"
)

func TestDeterministicSplitter(t *testing.T) {
	registry := NewConnectorRegistry()
	s := NewDeterministicSplitter(registry)

	tests := []struct {
		name       string
		utterance  string
		isValid    func(string) bool
		want       []string
	}{
		{
			name:      "Single intent, no connector",
			utterance: "Open Chrome",
			isValid:   func(chunk string) bool { return true },
			want:      []string{"Open Chrome"},
		},
		{
			name:      "Two intents separated by 'and'",
			utterance: "Open Chrome and search Google",
			isValid: func(chunk string) bool {
				return chunk == "Open Chrome" || chunk == "search Google"
			},
			want: []string{"Open Chrome", "search Google"},
		},
		{
			name:      "Negative case: 'and' within an entity, should not split",
			utterance: "Search for fish and chips",
			isValid: func(chunk string) bool {
				// "Search for fish" is not a valid goal in this context
				return chunk == "Search for fish and chips"
			},
			want: []string{"Search for fish and chips"},
		},
		{
			name:      "Three intents",
			utterance: "Create a reminder, take a note, and open Calendar",
			isValid: func(chunk string) bool {
				return chunk == "Create a reminder" || chunk == "take a note" || chunk == "open Calendar"
			},
			want: []string{"Create a reminder", "take a note", "open Calendar"},
		},
		{
			name:      "Fallback when split produces invalid goals",
			utterance: "Open Chrome and invalid text",
			isValid: func(chunk string) bool {
				return chunk == "Open Chrome" || chunk == "Open Chrome and invalid text"
			},
			// "Open Chrome" is valid, so it splits, and the remainder "invalid text" is just appended.
			want: []string{"Open Chrome", "invalid text"},
		},
		// Linguistic Regression Cases
		{
			name:      "Should split: Turn on WiFi and Bluetooth",
			utterance: "Turn on WiFi and Bluetooth",
			isValid: func(chunk string) bool {
				return chunk == "Turn on WiFi" || chunk == "Bluetooth" || chunk == "Turn on WiFi and Bluetooth"
			},
			want: []string{"Turn on WiFi", "Bluetooth"},
		},
		{
			name:      "Should split: Open Chrome then Spotify",
			utterance: "Open Chrome then Spotify",
			isValid: func(chunk string) bool {
				return chunk == "Open Chrome" || chunk == "Spotify" || chunk == "Open Chrome then Spotify"
			},
			want: []string{"Open Chrome", "Spotify"},
		},
		{
			name:      "Should split: Weather in Delhi and battery level",
			utterance: "Weather in Delhi and battery level",
			isValid: func(chunk string) bool {
				return chunk == "Weather in Delhi" || chunk == "battery level"
			},
			want: []string{"Weather in Delhi", "battery level"},
		},
		{
			name:      "Should NOT split: Fish and chips",
			utterance: "Order fish and chips",
			isValid: func(chunk string) bool {
				return chunk == "Order fish and chips"
			},
			want: []string{"Order fish and chips"},
		},
		{
			name:      "Should NOT split: Monday and Tuesday",
			// In practice, this would be masked by the orchestrator, but testing the splitter fallback if unmasked.
			utterance: "Schedule a meeting for Monday and Tuesday",
			isValid: func(chunk string) bool {
				return chunk == "Schedule a meeting for Monday and Tuesday"
			},
			want: []string{"Schedule a meeting for Monday and Tuesday"},
		},
		{
			name:      "Should NOT split: 3pm and 5pm",
			utterance: "Set an alarm for 3pm and 5pm",
			isValid: func(chunk string) bool {
				return chunk == "Set an alarm for 3pm and 5pm"
			},
			want: []string{"Set an alarm for 3pm and 5pm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Split(tt.utterance, tt.isValid)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkDeterministicSplitter_WorstCase(b *testing.B) {
	registry := NewConnectorRegistry()
	s := NewDeterministicSplitter(registry)

	// A worst case string with 10 connectors
	// Old O(2^N) would test 1024 combinations. New should test 10.
	utterance := "do A and do B and do C and do D and do E and do F and do G and do H and do I and do J and do K"
	
	// Validation function is fast
	isValid := func(chunk string) bool {
		return true
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Split(utterance, isValid)
	}
}
