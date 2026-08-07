package understanding

import (
	"regexp"
	"strings"
)

// GrammarDefinition represents an ordered collection of LanguageComponents.
// It acts as the explicit contract between Builders and the Compiler.
type GrammarDefinition struct {
	ID          string
	Intent      string
	Confidence  float64
	
	// Grouped components for analysis, certification, and standard sequential processing
	Prefixes    []LanguageComponent
	Concepts    []LanguageComponent
	Properties  []LanguageComponent
	Connectors  []LanguageComponent
	Slots       []LanguageComponent
	Terminators []LanguageComponent

	// Sequence explicitly defines the ordered sequence of components. 
	// If empty, the Compiler will use a standard concatenation of the groups above.
	Sequence    []LanguageComponent
}

// GrammarCompiler is responsible for translating a GrammarDefinition into a static PatternRule.
type GrammarCompiler struct {
	synonyms map[string][]string
}

// NewGrammarCompiler creates a new compiler with synonym awareness.
func NewGrammarCompiler(syns map[string][]string) *GrammarCompiler {
	if syns == nil {
		syns = DefaultSynonyms
	}
	return &GrammarCompiler{
		synonyms: syns,
	}
}

// Compile translates a GrammarDefinition into a *PatternRule exactly once.
func (c *GrammarCompiler) Compile(def GrammarDefinition) *PatternRule {
	var seq []LanguageComponent
	if len(def.Sequence) > 0 {
		seq = def.Sequence
	} else {
		// Standard order if not explicitly sequenced
		seq = append(seq, def.Prefixes...)
		seq = append(seq, def.Concepts...)
		seq = append(seq, def.Properties...)
		seq = append(seq, def.Connectors...)
		seq = append(seq, def.Slots...)
		seq = append(seq, def.Terminators...)
	}

	var patternBuilder strings.Builder
	patternBuilder.WriteString("^(?i)") // Case insensitive start

	for i, comp := range seq {
		part := c.compileComponent(comp)
		
		// Add appropriate spacing between components (except terminators and slots that might handle their own spacing)
		if i > 0 && comp.Category != "Terminator" {
			// In standard regex generation, we allow optional spacing
			// We handle explicit spacing inside compileComponent if needed, 
			// but a general \s* or \s+ helps join components.
			if comp.Category == "Slot" {
				patternBuilder.WriteString(`\s+`)
			} else {
				// If the component part already contains \s+, we don't necessarily need another, 
				// but adding \s* is safe for joining.
				// For purely deterministic behavior, we add \s+ unless it's optional.
				// Let's use `\s+` but make it optional `\s*` if components might be optional.
				patternBuilder.WriteString(`\s*`)
			}
		}
		patternBuilder.WriteString(part)
	}

	expr := patternBuilder.String()
	
	// Post-process to fix any double spaces or spacing before optional components if necessary
	expr = strings.ReplaceAll(expr, `\s*\s*`, `\s*`)
	
	return &PatternRule{
		id:      def.ID,
		pattern: regexp.MustCompile(expr),
		intent:  def.Intent,
		conf:    def.Confidence,
	}
}

func (c *GrammarCompiler) compileComponent(comp LanguageComponent) string {
	// The compiler converts semantic components into regex representation.
	
	val := comp.Value
	
	// Handle Slot categories dynamically based on value
	if comp.Category == "Slot" {
		name := comp.Name
		if name == "" {
			name = val
		}
		// Sanitize name for regex capture group
		name = strings.ReplaceAll(name, " ", "_")
		
		if val == "expression" {
			return `(?P<` + name + `>[\d\s\+\-\*\/\(\)\.]+)`
		} else if strings.HasPrefix(val, "operand") {
			return `(?P<` + name + `>\d+(?:\.\d+)?)`
		} else if val == "operator" {
			return `(?P<` + name + `>[\+\-\*\/])`
		} else if val == "data_text" {
			return `['"]?(?P<` + name + `>.*?)['"]?`
		} else if name == "Anaphora" {
			return `(?P<Anaphora>` + val + `)`
		}
		return `(?P<` + name + `>.+?)`
	}

	if comp.Category == "Terminator" {
		if val == "punctuation" {
			return `\s*[.!?]*$`
		}
		return val
	}

	// Escape standard values if they contain literal special characters,
	// though values like "what is|check" should be split by | and escaped individually.
	parts := strings.Split(val, "|")
	var escapedParts []string
	
	for _, p := range parts {
		// Check for synonyms if it's a single word or defined phrase
		if syns, exists := c.synonyms[p]; exists && len(syns) > 0 {
			allTerms := append([]string{p}, syns...)
			for j, term := range allTerms {
				allTerms[j] = regexp.QuoteMeta(term)
			}
			escapedParts = append(escapedParts, strings.Join(allTerms, "|"))
		} else {
			// Not in synonyms, just escape it
			escapedParts = append(escapedParts, regexp.QuoteMeta(p))
		}
	}

	// Join the parts
	joined := strings.Join(escapedParts, "|")
	if len(parts) > 1 || strings.Contains(joined, "|") {
		joined = "(?:" + joined + ")"
	}

	// Determine if optional based on Category conventions
	// (Compiler adds optional groups based on deterministic rules)
	isOptional := false
	if comp.Category == "Prefix" && comp.ID == "LANG_PREFIX_QUERY" {
		isOptional = true
	} else if comp.Category == "Connector" && (comp.Value == "the" || comp.Value == "my" || comp.Value == "a|an" || comp.Value == "the|my|a|an" || comp.Value == "for" || comp.Value == "called|named|titled") {
		isOptional = true
	} else if comp.Category == "SystemProperty" || comp.Category == "Property" {
		isOptional = true // Properties like "status" or "usage" are often optional in queries
	} else if comp.Category == "OptionalConcept" {
		isOptional = true
	}
	
	// Format with optional grouping if needed
	if isOptional {
		if strings.HasPrefix(joined, "(?:") && strings.HasSuffix(joined, ")") {
			return joined + "?"
		}
		return "(?:" + joined + ")?"
	}

	if !strings.HasPrefix(joined, "(?:") && strings.Contains(joined, "|") {
		joined = "(?:" + joined + ")"
	}

	if comp.Category == "ActionPrefix" {
		return `(?P<operation>` + strings.TrimPrefix(strings.TrimSuffix(joined, ")"), "(?:") + `)`
	}

	return joined
}
