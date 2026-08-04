package splitter

import (
	"reflect"
	"testing"
)

func TestDeterministicSplitter(t *testing.T) {
	s := NewDeterministicSplitter()

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
				// Mock valid goals
				return chunk == "Open Chrome" || chunk == "search Google"
			},
			want: []string{"Open Chrome", "search Google"},
		},
		{
			name:      "Negative case: 'and' within an entity, should not split",
			utterance: "Search for fish and chips",
			isValid: func(chunk string) bool {
				// "Search for fish" might be valid?
				// But "chips" is not a valid goal!
				// Only the full utterance is a valid goal.
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
			// "invalid text" is not a valid goal, so it falls back to full utterance
			want: []string{"Open Chrome and invalid text"},
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
