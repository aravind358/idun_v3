package understanding

import (
	"strings"
	"sync"
	"unicode"
)

// GrammarRule defines a single deterministic pattern or grammar matcher.
type GrammarRule interface {
	ID() string
	Match(norm NormalizedText) (intent string, slots []Slot, conf float64, matched bool)
}

// ExactKeywordRule matches when the cleaned utterance equals a specific keyword or phrase.
type ExactKeywordRule struct {
	id     string
	phrase string
	intent string
	conf   float64
}

// NewExactKeywordRule creates a rule matching an exact cleaned phrase.
func NewExactKeywordRule(id, phrase, intent string, conf float64) *ExactKeywordRule {
	return &ExactKeywordRule{
		id:     id,
		phrase: strings.ToLower(strings.TrimSpace(phrase)),
		intent: intent,
		conf:   conf,
	}
}

func (r *ExactKeywordRule) ID() string { return r.id }

func (r *ExactKeywordRule) Match(norm NormalizedText) (string, []Slot, float64, bool) {
	cleanNoPunct := strings.TrimFunc(norm.Cleaned, func(ru rune) bool {
		return unicode.IsPunct(ru) && ru != '_' && ru != '-' && ru != ':' && ru != '/'
	})
	cleanNoPunct = strings.TrimSpace(cleanNoPunct)
	if norm.Cleaned == r.phrase || cleanNoPunct == r.phrase {
		return r.intent, nil, r.conf, true
	}
	return "", nil, 0.0, false
}

// PrefixSlotRule matches when the cleaned utterance starts with a prefix phrase
// and captures the remainder as a named slot.
type PrefixSlotRule struct {
	id       string
	prefix   string
	intent   string
	slotName string
	conf     float64
}

// NewPrefixSlotRule creates a rule capturing remainder text after a prefix.
func NewPrefixSlotRule(id, prefix, intent, slotName string, conf float64) *PrefixSlotRule {
	return &PrefixSlotRule{
		id:       id,
		prefix:   strings.ToLower(strings.TrimSpace(prefix)) + " ",
		intent:   intent,
		slotName: slotName,
		conf:     conf,
	}
}

func (r *PrefixSlotRule) ID() string { return r.id }

func (r *PrefixSlotRule) Match(norm NormalizedText) (string, []Slot, float64, bool) {
	if strings.HasPrefix(norm.Cleaned, r.prefix) {
		val := strings.TrimSpace(strings.TrimPrefix(norm.Cleaned, r.prefix))
		if val != "" {
			slot := Slot{
				Name:        r.slotName,
				Value:       val,
				GroundingID: "slot-" + r.slotName,
				Confidence:  r.conf,
			}
			return r.intent, []Slot{slot}, r.conf, true
		}
	}
	cleanNoPunct := strings.TrimFunc(norm.Cleaned, func(ru rune) bool {
		return unicode.IsPunct(ru) && ru != '_' && ru != '-' && ru != ':' && ru != '/'
	})
	cleanNoPunct = strings.TrimSpace(cleanNoPunct)
	if strings.HasPrefix(cleanNoPunct, r.prefix) {
		val := strings.TrimSpace(strings.TrimPrefix(cleanNoPunct, r.prefix))
		if val != "" {
			slot := Slot{
				Name:        r.slotName,
				Value:       val,
				GroundingID: "slot-" + r.slotName,
				Confidence:  r.conf,
			}
			return r.intent, []Slot{slot}, r.conf, true
		}
	}
	return "", nil, 0.0, false
}

// GrammarSpecialist manages and evaluates registered deterministic grammar rules.
type GrammarSpecialist interface {
	RegisterRule(rule GrammarRule) error
	Evaluate(norm NormalizedText, boundSlots []Slot) (Hypothesis, bool)
}

// DefaultGrammarSpecialist implements thread-safe deterministic rule evaluation.
type DefaultGrammarSpecialist struct {
	mu    sync.RWMutex
	rules []GrammarRule
}

// NewDefaultGrammarSpecialist creates a GrammarSpecialist pre-loaded with built-in rules.
func NewDefaultGrammarSpecialist() *DefaultGrammarSpecialist {
	g := &DefaultGrammarSpecialist{}
	_ = g.RegisterRule(NewExactKeywordRule("rule.status", "status", "query_status", 0.99))
	_ = g.RegisterRule(NewExactKeywordRule("rule.cancel", "cancel", "cancel_action", 0.99))
	_ = g.RegisterRule(NewPrefixSlotRule("rule.alarm", "set alarm for", "set_alarm", "time", 0.96))
	_ = g.RegisterRule(NewPrefixSlotRule("rule.weather", "weather in", "query_weather", "city", 0.95))
	_ = g.RegisterRule(NewExactKeywordRule("rule.hello", "hello", "greet_user", 0.98))
	_ = g.RegisterRule(NewExactKeywordRule("rule.hello_idun", "hello idun", "greet_user", 0.98))
	_ = g.RegisterRule(NewExactKeywordRule("rule.hi", "hi", "greet_user", 0.98))
	_ = g.RegisterRule(NewExactKeywordRule("rule.who", "who are you", "query_identity", 0.98))
	_ = g.RegisterRule(NewExactKeywordRule("rule.how", "how are you", "query_wellbeing", 0.98))
	_ = g.RegisterRule(NewExactKeywordRule("rule.goodbye", "goodbye", "farewell_user", 0.98))
	return g
}

// RegisterRule appends a deterministic rule to the specialist.
func (g *DefaultGrammarSpecialist) RegisterRule(rule GrammarRule) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rules = append(g.rules, rule)
	return nil
}

// Evaluate checks the normalized text against registered rules in priority order.
func (g *DefaultGrammarSpecialist) Evaluate(norm NormalizedText, boundSlots []Slot) (Hypothesis, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, rule := range g.rules {
		intent, extractedSlots, conf, matched := rule.Match(norm)
		if matched {
			combinedSlots := make([]Slot, 0, len(boundSlots)+len(extractedSlots))
			combinedSlots = append(combinedSlots, boundSlots...)
			combinedSlots = append(combinedSlots, extractedSlots...)

			hyp := Hypothesis{
				Intent:               intent,
				CalibratedConfidence: conf,
				SourceLayer:          LayerReflexiveGrammar,
				Slots:                combinedSlots,
			}
			return hyp, true
		}
	}
	return Hypothesis{}, false
}

// Ensure DefaultGrammarSpecialist implements GrammarSpecialist.
var _ GrammarSpecialist = (*DefaultGrammarSpecialist)(nil)
