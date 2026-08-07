package understanding

// LoadContextRules registers the deterministically compiled Context rules
// into the GrammarSpecialist.
func LoadContextRules(g GrammarSpecialist, lib *GrammarBuilderLibrary) {

	// context_action (e.g. "Delete it", "Open that")
	// Since Action prefixes have different forms, we can use a generic BuildRule
	// or define a generic one using a combined prefix. We'll use a broad ActionPrefix combination.
	// We'll create a custom component just for this or rely on a standard combination.
	// To keep it simple, we'll combine common action prefixes into a new component or just use one.
	// Actually, let's create a temporary combined action prefix for this, or use the builder directly with an inline component if needed.
	// Wait, the U6 principles say "Reuse before creation", but we want to catch ANY action on an anaphora.
	// We can use a combined action prefix.
	allActionsPrefix := LanguageComponent{
		ID:       "LANG_PREFIX_ALL_ACTIONS",
		Name:     "All Actions",
		Value:    "delete|remove|open|show|read|create|make|set|move|copy|rename|start|run",
		Category: "ActionPrefix",
	}

	g.RegisterRule(lib.BuildRule("rule.context.action", "context_action", 0.95,
		allActionsPrefix,
		ConceptAnaphora,
		TerminatorPunctuation,
	))

	// context_ellipsis.concept (e.g. "what about battery")
	g.RegisterRule(lib.BuildRule("rule.context.ellipsis.concept", "context_ellipsis", 0.95,
		ConnectorEllipsis,
		LanguageComponent{Category: "Slot", Value: "concept"}, // capture the rest as a generic slot
		TerminatorPunctuation,
	))
	// Wait, the user prompt said:
	// "what about battery" -> Intent: context_ellipsis, Target: "battery"
	// To capture "battery", we can use a Slot component.

	// confirmation (e.g. "yes", "do it")
	g.RegisterRule(lib.BuildRule("rule.context.confirmation", "confirmation", 0.95,
		ConceptAffirmation,
		TerminatorPunctuation,
	))

	// negation (e.g. "no", "cancel")
	g.RegisterRule(lib.BuildRule("rule.context.negation", "negation", 0.95,
		ConceptNegation,
		TerminatorPunctuation,
	))
}
