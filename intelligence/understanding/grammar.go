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

// NewPatternRule creates a rule matching a regex pattern.
func NewPatternRule(id, expr, intent string, conf float64) *PatternRule {
	return &PatternRule{
		id:      id,
		pattern: regexp.MustCompile(expandPattern(expr)),
		intent:  intent,
		conf:    conf,
	}
}

func expandPattern(expr string) string {
	for k, v := range DefaultSynonyms {
		synGroup := k
		if len(v) > 0 {
			synGroup = k + "|" + strings.Join(v, "|")
		}
		expr = strings.ReplaceAll(expr, "{"+k+"}", "(?:"+synGroup+")")
	}
	return expr
}


func (r *PatternRule) ID() string { return r.id }

func (r *PatternRule) Match(norm NormalizedText) (string, []Slot, float64, bool) {
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
	_ = g.RegisterRule(NewPatternRule("rule.status", `^(?i)(?:{what_is}\s+)?(?:the\s+)?{status}$`, "query_status", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.cancel", `^(?i){cancel}$`, "cancel_action", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.alarm", `^(?i)(?:{set}\s+)?{alarm}\s+(?:for\s+)?(?P<time>.+)$`, "set_alarm", 0.96))
	_ = g.RegisterRule(NewPatternRule("rule.weather_city", `^(?i)(?:{what_is}\s+)?(?:the\s+)?{weather}\s+(?:in\s+)?(?P<city>.+)$`, "query_weather", 0.95))
	_ = g.RegisterRule(NewPatternRule("rule.weather", `^(?i)(?:{what_is}\s+)?(?:the\s+)?{weather}$`, "query_weather", 0.90))
	_ = g.RegisterRule(NewPatternRule("rule.time", `^(?i)(?:{what_is}\s+)?(?:the\s+|current\s+)?time(?:\s+is\s+it|\s+now)?\s*[.!?]*$`, "query_time", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.date", `^(?i)(?:{what_is}\s+)?(?:today'?s\s+|current\s+|the\s+)?(?:date|day)(?:\s+is\s+today)?\s*[.!?]*$`, "query_date", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.hello", `^(?i){hello}(?:\s+idun)?$`, "greet_user", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.calculate", `^(?i)(?:calculate\s+|what is\s+)?(?P<operand1>\d+(?:\.\d+)?)\s*(?P<operator>[\+\-\*\/])\s*(?P<operand2>\d+(?:\.\d+)?)\s*[.!?]*$`, "calculate", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.calculate.complex", `^(?i)(?:calculate\s+|what is\s+)?(?P<expression>[\d\s\+\-\*\/\(\)\.]+)\s*[.!?]*$`, "calculate", 0.95))
	_ = g.RegisterRule(NewPatternRule("rule.reminder.full", `^(?i)(?P<operation>remind)\s+(?P<person>\w+)\s+(?P<time>.+?)\s+to\s+(?P<task>.+?)\s*[.!?]*$`, "create_reminder", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.reminder.simple", `^(?i)(?P<operation>remind)\s+(?P<person>\w+)\s+(?P<time>[^t].*?)\s*[.!?]*$`, "create_reminder", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.reminder.create", `^(?i)(?P<operation>create)\s+(?:a\s+)?reminder\s+(?:for\s+)?(?P<time>.+?)\s*[.!?]*$`, "create_reminder", 0.98))

	// Notes Grammar
	_ = g.RegisterRule(NewPatternRule("rule.note.create.full", `^(?i)(?P<operation>set|create|save|take)\s+a\s+note\s+(?:titled|called)\s+(?P<title>.+?)\s+saying\s+(?P<content>.+?)\s*[.!?]*$`, "manage_notes", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.note.create.content1", `^(?i)(?P<operation>take)\s+a\s+note\s+(?:to|saying)\s+(?P<content>.+?)\s*[.!?]*$`, "manage_notes", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.note.create.content2", `^(?i)(?P<operation>note)\s+that\s+(?P<content>.+?)\s*[.!?]*$`, "manage_notes", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.note.create.content3", `^(?i)(?P<operation>save)\s+a\s+note\s+saying\s+(?P<content>.+?)\s*[.!?]*$`, "manage_notes", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.note.create.title", `^(?i)(?P<operation>set|create)\s+a\s+note\s+(?:titled|called)\s+(?P<title>.+?)\s*[.!?]*$`, "manage_notes", 0.98))
	
	_ = g.RegisterRule(NewPatternRule("rule.note.list1", `^(?i)(?P<operation>list|show)\s+my\s+notes\s*[.!?]*$`, "manage_notes", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.note.list2", `^(?i)(?P<operation>what)\s+notes\s+do\s+i\s+have\s*[.!?]*$`, "manage_notes", 0.99))

	_ = g.RegisterRule(NewPatternRule("rule.note.read1", `^(?i)(?P<operation>read|open|show)\s+my\s+(?P<title>.+?)\s+note\s*[.!?]*$`, "manage_notes", 0.97))
	_ = g.RegisterRule(NewPatternRule("rule.note.read2", `^(?i)(?P<operation>show)\s+my\s+note\s+(?:called|titled)\s+(?P<title>.+?)\s*[.!?]*$`, "manage_notes", 0.97))
	
	_ = g.RegisterRule(NewPatternRule("rule.note.delete1", `^(?i)(?P<operation>delete|remove)\s+my\s+(?P<title>.+?)\s+note\s*[.!?]*$`, "manage_notes", 0.97))
	_ = g.RegisterRule(NewPatternRule("rule.note.delete2", `^(?i)(?P<operation>delete|remove)\s+(?:the|my)\s+note\s+(?:called|titled)\s+(?P<title>.+?)\s*[.!?]*$`, "manage_notes", 0.97))

	// File Grammar
	_ = g.RegisterRule(NewPatternRule("rule.file.move", `^(?i)(?P<operation>move)\s+(?:file\s+)?(?P<source>.+?)\s+to\s+(?P<destination>.+?)\s*[.!?]*$`, "file_operation", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.file.copy", `^(?i)(?P<operation>copy)\s+(?:file\s+)?(?P<source>.+?)\s+to\s+(?P<destination>.+?)\s*[.!?]*$`, "file_operation", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.file.copy_only", `^(?i)(?P<operation>copy)\s+(?:file\s+)?(?P<source>.+?)\s*[.!?]*$`, "file_operation", 0.97))
	_ = g.RegisterRule(NewPatternRule("rule.file.rename", `^(?i)(?P<operation>rename)\s+(?:file\s+)?(?P<source>.+?)\s+to\s+(?P<destination>.+?)\s*[.!?]*$`, "file_operation", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.file.open", `^(?i)(?P<operation>open|read)\s+(?:file\s+)?(?P<filename>.+?)\s*[.!?]*$`, "file_operation", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.file.delete", `^(?i)(?P<operation>delete|remove)\s+(?:file\s+)?(?P<filename>.+?)\s*[.!?]*$`, "file_operation", 0.98))

	// Dangerous Ops Grammar
	_ = g.RegisterRule(NewPatternRule("rule.file.dangerous", `^(?i)(?P<operation>format\s+drive|erase\s+disk)$`, "file_operation", 0.99))

	// Directory Grammar
	_ = g.RegisterRule(NewPatternRule("rule.dir.create1", `^(?i)(?P<operation>create|make)\s+(?:a\s+)?(?:directory|folder)\s+(?:called\s+|named\s+)?(?P<directory>.+?)\s*[.!?]*$`, "create_directory", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.dir.create2", `^(?i)(?P<operation>mkdir)\s+(?P<directory>.+?)\s*[.!?]*$`, "create_directory", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.dir.list1", `^(?i)(?P<operation>list|show)\s+(?:all\s+)?(?:files|contents)(?:\s+in\s+|\s+of\s+)(?P<directory>.+?)\s*[.!?]*$`, "list_files", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.dir.list_simple", `^(?i)(?P<operation>list)\s+(?P<directory>.+?)\s*[.!?]*$`, "list_files", 0.97))
	_ = g.RegisterRule(NewPatternRule("rule.dir.list2", `^(?i)(?P<operation>list|show)\s+(?:all\s+)?files\s*[.!?]*$`, "list_files", 0.98))

	// System Information Grammar
	_ = g.RegisterRule(NewPatternRule("rule.sys.battery1", `^(?i)(?:what\s+is\s+(?:the\s+)?|check\s+(?:the\s+)?|query\s+)?battery(?:(?:\s+level|\s+percentage|\s+status))?\s*[.!?]*$`, "query_battery", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.sys.cpu1", `^(?i)(?:what\s+is\s+(?:the\s+)?|check\s+(?:the\s+)?|query\s+)?cpu(?:\s+usage)?\s*[.!?]*$`, "query_cpu", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.sys.mem1", `^(?i)(?:what\s+is\s+(?:the\s+)?|check\s+(?:the\s+)?|query\s+)?memory(?:\s+usage)?\s*[.!?]*$`, "query_memory", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.sys.disk1", `^(?i)(?:what\s+is\s+(?:the\s+)?|check\s+(?:the\s+)?|query\s+)?disk(?:\s+usage|\s+space)?\s*[.!?]*$`, "query_disk", 0.98))

	// System Control Grammar
	_ = g.RegisterRule(NewPatternRule("rule.sys.shutdown1", `^(?i)(?P<operation>shut\s+down|shutdown|turn\s+off)\s+(?:the\s+)?(?:computer|system|pc|machine)\s*[.!?]*$`, "system_shutdown", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.sys.shutdown2", `^(?i)(?P<operation>shut\s+down|shutdown)\s*[.!?]*$`, "system_shutdown", 0.95))
	_ = g.RegisterRule(NewPatternRule("rule.sys.restart1", `^(?i)(?P<operation>restart|reboot)\s+(?:the\s+)?(?:computer|system|pc|machine)\s*[.!?]*$`, "system_restart", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.sys.restart2", `^(?i)(?P<operation>restart|reboot)\s*[.!?]*$`, "system_restart", 0.95))
	_ = g.RegisterRule(NewPatternRule("rule.sys.lock1", `^(?i)(?P<operation>lock)\s+(?:the\s+)?(?:computer|system|pc|machine|screen)\s*[.!?]*$`, "system_lock", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.sys.lock2", `^(?i)(?P<operation>lock)\s*[.!?]*$`, "system_lock", 0.95))

	// Dangerous System Grammar (Mapped to System Operations for PermissionPolicy rejection)
	_ = g.RegisterRule(NewPatternRule("rule.sys.destroy", `^(?i)(?P<operation>destroy|kill|wipe)\s+(?:the\s+)?(?:computer|system|pc|machine|windows|all\s+processes|disk|my\s+disk)\s*[.!?]*$`, "system_shutdown", 0.99))
	_ = g.RegisterRule(NewPatternRule("rule.sys.shutdown_all", `^(?i)(?P<operation>shut\s+everything\s+down)\s*[.!?]*$`, "system_shutdown", 0.99))

	_ = g.RegisterRule(NewPatternRule("rule.who", `^(?i)(?:who|what)\s+are\s+you$`, "query_identity", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.how", `^(?i)how\s+are\s+you$`, "query_wellbeing", 0.98))
	_ = g.RegisterRule(NewPatternRule("rule.goodbye", `^(?i){goodbye}$`, "farewell_user", 0.98))
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
