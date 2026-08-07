package understanding

// LoadTimeAndWeatherRules registers the deterministically compiled Time and Weather capabilities
// into the GrammarSpecialist.
func LoadTimeAndWeatherRules(g GrammarSpecialist, lib *GrammarBuilderLibrary) {
	// -------------------------------------------------------------------------
	// Time Rules
	// -------------------------------------------------------------------------
	g.RegisterRule(lib.BuildRule("rule.time", "query_time", 0.99,
		QueryPrefix,
		ConnectorThe,
		ConceptTime,
		TerminatorPunctuation,
	))

	// -------------------------------------------------------------------------
	// Date Rules
	// -------------------------------------------------------------------------
	g.RegisterRule(lib.BuildRule("rule.date", "query_date", 0.99,
		QueryPrefix,
		ConnectorThe,
		ConceptDate,
		TerminatorPunctuation,
	))

	// -------------------------------------------------------------------------
	// Weather Rules
	// -------------------------------------------------------------------------

	// rule.weather.simple (e.g., "what is the weather")
	g.RegisterRule(lib.BuildRule("rule.weather.simple", "query_weather", 0.99,
		QueryPrefix,
		ConnectorThe,
		ConceptWeather,
		TerminatorPunctuation,
	))

	// rule.weather.complex (e.g., "what is the weather in <location>")
	g.RegisterRule(lib.BuildRule("rule.weather.complex", "query_weather", 0.99,
		QueryPrefix,
		ConnectorThe,
		ConceptWeather,
		ConnectorIn,
		SlotLocation,
		TerminatorPunctuation,
	))

	// rule.weather.outside (e.g., "what is it like outside")
	// Since "outside" is weather, we map it directly.
	g.RegisterRule(lib.BuildRule("rule.weather.outside", "query_weather", 0.99,
		QueryPrefix,
		LanguageComponent{Category: "Connector", Value: "it like"},
		LanguageComponent{Category: "Concept", Value: "outside"},
		TerminatorPunctuation,
	))

	// -------------------------------------------------------------------------
	// Alarm Rules
	// -------------------------------------------------------------------------
	g.RegisterRule(lib.BuildRule("rule.alarm", "set_alarm", 0.96,
		CreateActionPrefix, // "set|create|make|add"
		ConnectorThe,       // optional "the|a|an"
		ConceptAlarm,       // "alarm|timer|alert"
		ConnectorFor,       // optional "for"
		SlotTime,           // (?P<time>.+)
		TerminatorPunctuation,
	))
}
