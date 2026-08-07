package understanding

func LoadMetaRules(g *DefaultGrammarSpecialist, lib *GrammarBuilderLibrary) {
	// -------------------------------------------------------------------------
	// Status Rules
	// -------------------------------------------------------------------------
	
	// rule.status
	g.RegisterRule(lib.BuildRule("rule.status", "query_status", 0.95,
		QueryPrefix, // optional
		ConnectorThe, // optional
		LanguageComponent{Category: "Property", Value: "status|system status|current status"},
		TerminatorPunctuation,
	))

	// -------------------------------------------------------------------------
	// Meta / Meta-Conversational Rules
	// -------------------------------------------------------------------------

	// rule.cancel
	g.RegisterRule(lib.BuildRule("rule.cancel", "cancel_action", 0.98,
		LanguageComponent{Category: "ActionPrefix", Value: "cancel|stop|abort|nevermind|never mind|quit|exit|terminate|halt"},
		TerminatorPunctuation,
	))

	// rule.hello
	g.RegisterRule(lib.BuildRule("rule.hello", "greet_user", 0.98,
		LanguageComponent{Category: "KnowledgeConcept", Value: "hello"},
		LanguageComponent{Category: "OptionalConcept", Value: "idun"},
		TerminatorPunctuation,
	))

	// rule.goodbye
	g.RegisterRule(lib.BuildRule("rule.goodbye", "farewell_user", 0.98,
		LanguageComponent{Category: "KnowledgeConcept", Value: "goodbye"},
		TerminatorPunctuation,
	))

	// rule.who
	g.RegisterRule(lib.BuildRule("rule.who", "query_identity", 0.98,
		LanguageComponent{Category: "RequiredConnector", Value: "who|what"},
		LanguageComponent{Category: "RequiredConnector", Value: "are"},
		LanguageComponent{Category: "RequiredConnector", Value: "you"},
		TerminatorPunctuation,
	))

	// rule.how
	g.RegisterRule(lib.BuildRule("rule.how", "query_wellbeing", 0.98,
		LanguageComponent{Category: "RequiredConnector", Value: "how"},
		LanguageComponent{Category: "RequiredConnector", Value: "are"},
		LanguageComponent{Category: "RequiredConnector", Value: "you"},
		TerminatorPunctuation,
	))

	// rule.system.help
	g.RegisterRule(lib.BuildRule("rule.system.help", "system_help", 0.99,
		LanguageComponent{Category: "ActionPrefix", Value: "help|what can you do|how to use"},
		TerminatorPunctuation,
	))

	// -------------------------------------------------------------------------
	// Math / Calculator
	// -------------------------------------------------------------------------

	// rule.calculate
	calcPrefixCombine := LanguageComponent{Category: "OptionalConcept", Value: "calculate|what_is"}
	g.RegisterRule(lib.BuildRule("rule.calculate", "calculate", 0.99,
		calcPrefixCombine,
		LanguageComponent{Category: "Slot", Value: "operand1"},
		LanguageComponent{Category: "Slot", Value: "operator"},
		LanguageComponent{Category: "Slot", Value: "operand2"},
		TerminatorPunctuation,
	))

	// rule.calculate.complex
	g.RegisterRule(lib.BuildRule("rule.calculate.complex", "calculate", 0.95,
		calcPrefixCombine,
		LanguageComponent{Category: "Slot", Value: "expression"},
		TerminatorPunctuation,
	))
}
