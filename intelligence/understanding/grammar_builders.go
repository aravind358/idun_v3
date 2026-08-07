package understanding

// GrammarBuilderLibrary defines the reusable language patterns
// independent of any specific capability (Battery, Weather, etc).
type GrammarBuilderLibrary struct {
	compiler *GrammarCompiler
}

// NewGrammarBuilderLibrary creates a new builder library connected to the compiler.
func NewGrammarBuilderLibrary(compiler *GrammarCompiler) *GrammarBuilderLibrary {
	return &GrammarBuilderLibrary{
		compiler: compiler,
	}
}

// BuildRule is a generic builder that creates a rule from a sequence of components.
// E.g., action + connector + concept + slot + terminator
func (b *GrammarBuilderLibrary) BuildRule(id, intent string, conf float64, sequence ...LanguageComponent) *PatternRule {
	def := GrammarDefinition{
		ID:         id,
		Intent:     intent,
		Confidence: conf,
		Sequence:   sequence,
	}
	return b.compiler.Compile(def)
}

// BuildSimpleQuery creates a rule for querying a concept without properties.
// E.g., "what is the time" -> QueryPrefix + The + Time + Terminator
func (b *GrammarBuilderLibrary) BuildSimpleQuery(id, intent string, conf float64, concept LanguageComponent) *PatternRule {
	def := GrammarDefinition{
		ID:         id,
		Intent:     intent,
		Confidence: conf,
		Sequence: []LanguageComponent{
			QueryPrefix,
			ConnectorThe,
			concept,
			TerminatorPunctuation,
		},
	}
	return b.compiler.Compile(def)
}

// BuildObjectPropertyQuery creates a rule for querying a concept's property.
// E.g., "what is the battery status" -> QueryPrefix + The + Battery + Status + Terminator
func (b *GrammarBuilderLibrary) BuildObjectPropertyQuery(id, intent string, conf float64, concept, property LanguageComponent) *PatternRule {
	def := GrammarDefinition{
		ID:         id,
		Intent:     intent,
		Confidence: conf,
		Sequence: []LanguageComponent{
			QueryPrefix,
			ConnectorThe,
			concept,
			property,
			TerminatorPunctuation,
		},
	}
	return b.compiler.Compile(def)
}

// BuildObjectAction creates a rule for performing an action on an object.
// E.g., "open file foo" -> ActionPrefix + The + Object + Slot + Terminator
func (b *GrammarBuilderLibrary) BuildObjectAction(id, intent string, conf float64, action, concept, targetSlot LanguageComponent) *PatternRule {
	def := GrammarDefinition{
		ID:         id,
		Intent:     intent,
		Confidence: conf,
		Sequence: []LanguageComponent{
			action,
			ConnectorThe,
			concept,
			targetSlot,
			TerminatorPunctuation,
		},
	}
	return b.compiler.Compile(def)
}

// BuildSourceDestinationAction creates a rule for actions involving movement.
// E.g., "move file X to Y" -> Action + Concept + SourceSlot + To + DestinationSlot + Terminator
func (b *GrammarBuilderLibrary) BuildSourceDestinationAction(id, intent string, conf float64, action, concept, sourceSlot, destSlot LanguageComponent) *PatternRule {
	def := GrammarDefinition{
		ID:         id,
		Intent:     intent,
		Confidence: conf,
		Sequence: []LanguageComponent{
			action,
			ConnectorThe,
			concept,
			sourceSlot,
			ConnectorTo,
			destSlot,
			TerminatorPunctuation,
		},
	}
	return b.compiler.Compile(def)
}

// BuildSlotQuery creates a rule based primarily on an action prefix and slots.
// E.g., "delete note X" -> Action + Concept + Slot
func (b *GrammarBuilderLibrary) BuildSlotQuery(id, intent string, conf float64, action, concept, slot LanguageComponent) *PatternRule {
	def := GrammarDefinition{
		ID:         id,
		Intent:     intent,
		Confidence: conf,
		Sequence: []LanguageComponent{
			action,
			ConnectorThe,
			concept,
			ConnectorCalled,
			slot,
			TerminatorPunctuation,
		},
	}
	return b.compiler.Compile(def)
}
