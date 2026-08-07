package understanding

import (
	"strings"
	"sync"
)

// NeuralSpecialist defines the abstraction for local probabilistic neural classification,
// slot prediction, and hypothesis generation without external AI models or network calls.
type NeuralSpecialist interface {
	Evaluate(norm NormalizedText) ([]Hypothesis, error)
}

// PatternClassifierRule represents a probabilistic classification pattern for the local neural specialist.
type PatternClassifierRule struct {
	Keywords []string
	Intent   string
	BaseConf float64
	SlotName string
}

// DefaultNeuralSpecialist implements a local probabilistic intent classifier abstraction.
type DefaultNeuralSpecialist struct {
	mu       sync.RWMutex
	patterns []PatternClassifierRule
}

// NewDefaultNeuralSpecialist creates a local probabilistic neural specialist abstraction.
func NewDefaultNeuralSpecialist() *DefaultNeuralSpecialist {
	n := &DefaultNeuralSpecialist{}
	n.patterns = []PatternClassifierRule{
		{
			Keywords: []string{"reschedule", "move", "postpone", "meeting"},
			Intent:   "reschedule_meeting",
			BaseConf: 0.85,
			SlotName: "target_event",
		},
		{
			Keywords: []string{"cancel", "delete", "drop", "meeting"},
			Intent:   "cancel_meeting",
			BaseConf: 0.82,
			SlotName: "target_event",
		},
		{
			Keywords: []string{"flight", "book", "ticket", "airline"},
			Intent:   "book_flight",
			BaseConf: 0.88,
			SlotName: "destination",
		},
		{
			Keywords: []string{"weather", "forecast", "temperature", "rain"},
			Intent:   "query_weather",
			BaseConf: 0.84,
			SlotName: "city",
		},
		{
			Keywords: []string{"hello", "hi", "greetings", "hey"},
			Intent:   "greet_user",
			BaseConf: 0.90,
		},
		{
			Keywords: []string{"who", "identity", "what are you"},
			Intent:   "query_identity",
			BaseConf: 0.88,
		},
		{
			Keywords: []string{"how are you", "wellbeing"},
			Intent:   "query_wellbeing",
			BaseConf: 0.86,
		},
		{
			Keywords: []string{"goodbye", "bye", "farewell"},
			Intent:   "farewell_user",
			BaseConf: 0.90,
		},
	}
	return n
}

// RegisterPattern registers a classification pattern on the neural specialist.
func (n *DefaultNeuralSpecialist) RegisterPattern(rule PatternClassifierRule) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.patterns = append(n.patterns, rule)
}

// Evaluate concurrently evaluates probabilistic hypotheses against the input utterance.
func (n *DefaultNeuralSpecialist) Evaluate(norm NormalizedText) ([]Hypothesis, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var hyps []Hypothesis

	for _, pat := range n.patterns {
		matchCount := 0
		for _, kw := range pat.Keywords {
			if strings.Contains(norm.Cleaned, kw) {
				matchCount++
			}
		}
		if matchCount > 0 {
			// Compute uncalibrated raw confidence scaled by keyword overlap
			conf := pat.BaseConf * (0.8 + 0.2*float64(matchCount)/float64(len(pat.Keywords)))
			if conf > 0.98 {
				conf = 0.98
			}

			var slots []Slot

			hyps = append(hyps, Hypothesis{
				Intent:               pat.Intent,
				CalibratedConfidence: conf,
				SourceLayer:          LayerNeuralClassifier,
				Slots:                slots,
			})
		}
	}

	return hyps, nil
}

// Ensure DefaultNeuralSpecialist implements NeuralSpecialist.
var _ NeuralSpecialist = (*DefaultNeuralSpecialist)(nil)
