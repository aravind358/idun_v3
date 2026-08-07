package understanding

import (
	"regexp"
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

// PatternRule matches when the cleaned utterance matches a regex pattern.
type PatternRule struct {
	id      string
	pattern *regexp.Regexp
	intent  string
	conf    float64
}



func (r *PatternRule) Pattern() string {
	return r.pattern.String()
}

func (r *PatternRule) Intent() string { return r.id }

func (r *PatternRule) ID() string { return r.id }

func (r *PatternRule) PatternString() string {
	if r.pattern != nil {
		return r.pattern.String()
	}
	return ""
}

func (r *PatternRule) Match(norm NormalizedText) (string, []Slot, float64, bool) {
	if norm.Cleaned == "" {
		return "", nil, 0, false
	}
	match := r.pattern.FindStringSubmatch(norm.Cleaned)
	if match == nil {
		return "", nil, 0.0, false
	}

	var slots []Slot
	names := r.pattern.SubexpNames()
	for i, name := range names {
		if i != 0 && name != "" && match[i] != "" {
			slots = append(slots, Slot{
				Name:        name,
				Value:       match[i],
				GroundingID: "slot-" + name,
				Confidence:  r.conf,
			})
		}
	}
	return r.intent, slots, r.conf, true
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


// GrammarSpecialist manages and evaluates registered deterministic grammar rules.
type GrammarSpecialist interface {
	RegisterRule(rule GrammarRule) error
	Evaluate(norm NormalizedText) (Hypothesis, bool)
}

// DefaultGrammarSpecialist implements thread-safe deterministic rule evaluation.
type DefaultGrammarSpecialist struct {
	mu    sync.RWMutex
	rules []GrammarRule
}

func (g *DefaultGrammarSpecialist) GetRules() []GrammarRule {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.rules
}

// NewDefaultGrammarSpecialist creates a GrammarSpecialist pre-loaded with built-in rules.
func NewDefaultGrammarSpecialist() *DefaultGrammarSpecialist {
	g := &DefaultGrammarSpecialist{}
	compiler := NewGrammarCompiler(DefaultSynonyms)
	lib := NewGrammarBuilderLibrary(compiler)
	
	// Wave 1: System Rules
	LoadSystemRules(g, lib)
	LoadContextRules(g, lib)

	// --- Wave 2: Time & Weather ---
	LoadTimeAndWeatherRules(g, lib)

	// --- Wave 4: Notes (More specific, must be before Files) ---
	LoadNoteRules(g, lib)

	// --- Wave 3: Files (More general) ---
	LoadFileRules(g, lib)

	// --- Wave 5: Meta (Fallback and system rules) ---
	LoadMetaRules(g, lib)

	// Default/Fallback
	g.RegisterRule(lib.BuildRule("rule.default", "unresolved_intent", 0.10, FallbackMatchAll))

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
func (g *DefaultGrammarSpecialist) Evaluate(norm NormalizedText) (Hypothesis, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, rule := range g.rules {
		intent, extractedSlots, conf, matched := rule.Match(norm)
		if matched {
			hyp := Hypothesis{
				Intent:               intent,
				CalibratedConfidence: conf,
				SourceLayer:          LayerReflexiveGrammar,
				Slots:                extractedSlots,
			}
			return hyp, true
		}
	}
	return Hypothesis{}, false
}

// Ensure DefaultGrammarSpecialist implements GrammarSpecialist.
var _ GrammarSpecialist = (*DefaultGrammarSpecialist)(nil)
