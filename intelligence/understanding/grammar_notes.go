package understanding

// LoadNoteRules registers all deterministic reminder and note rules.
func LoadNoteRules(g *DefaultGrammarSpecialist, lib *GrammarBuilderLibrary) {

	// -------------------------------------------------------------------------
	// Reminder Rules
	// -------------------------------------------------------------------------

	// rule.reminder.full
	g.RegisterRule(lib.BuildRule("rule.reminder.full", "create_reminder", 0.99,
		RemindActionPrefix,
		SlotPerson,
		SlotTime,
		ConnectorTo,
		SlotTask,
		TerminatorPunctuation,
	))

	// rule.reminder.simple
	// The compiler translates SlotTime to (?P<time>.+?) which acts exactly like [^t].*? since .full has higher confidence and picks off "to task".
	g.RegisterRule(lib.BuildRule("rule.reminder.simple", "create_reminder", 0.98,
		RemindActionPrefix,
		SlotPerson,
		SlotTime,
		TerminatorPunctuation,
	))

	// rule.reminder.create
	g.RegisterRule(lib.BuildRule("rule.reminder.create", "create_reminder", 0.98,
		CreateActionPrefix,
		ConnectorA, // optional via compiler
		ConceptReminder,
		ConnectorFor, // optional via compiler
		SlotTime,
		TerminatorPunctuation,
	))

	// -------------------------------------------------------------------------
	// Notes Rules
	// -------------------------------------------------------------------------

	noteCreatePrefix := LanguageComponent{
		ID:       "LANG_PREFIX_NOTE_CREATE",
		Name:     "Note Create Prefix",
		Value:    "set|create|save|take",
		Category: "ActionPrefix",
	}

	// rule.note.create.full
	g.RegisterRule(lib.BuildRule("rule.note.create.full", "manage_notes", 0.99,
		noteCreatePrefix,
		ConnectorA,
		ConceptNote,
		ConnectorCalled, // optional via compiler
		SlotTitle,
		ConnectorSaying,
		SlotContent,
		TerminatorPunctuation,
	))

	// rule.note.create.content1
	g.RegisterRule(lib.BuildRule("rule.note.create.content1", "manage_notes", 0.98,
		TakeActionPrefix,
		ConnectorA,
		ConceptNote,
		ConnectorToOrSaying,
		SlotContent,
		TerminatorPunctuation,
	))

	// rule.note.create.content2
	g.RegisterRule(lib.BuildRule("rule.note.create.content2", "manage_notes", 0.98,
		NoteActionPrefix,
		ConnectorThat,
		SlotContent,
		TerminatorPunctuation,
	))

	// rule.note.create.content3
	g.RegisterRule(lib.BuildRule("rule.note.create.content3", "manage_notes", 0.98,
		SaveActionPrefix,
		ConnectorA,
		ConceptNote,
		ConnectorSaying,
		SlotContent,
		TerminatorPunctuation,
	))

	// rule.note.create.title
	g.RegisterRule(lib.BuildRule("rule.note.create.title", "manage_notes", 0.98,
		CreateActionPrefix,
		ConnectorA,
		ConceptNote,
		ConnectorCalled, // optional via compiler
		SlotTitle,
		TerminatorPunctuation,
	))

	// rule.note.list1
	g.RegisterRule(lib.BuildRule("rule.note.list1", "manage_notes", 0.99,
		ListActionPrefix,
		ConnectorMy, // optional
		ConceptNote,
		TerminatorPunctuation,
	))

	// rule.note.list2
	g.RegisterRule(lib.BuildRule("rule.note.list2", "manage_notes", 0.99,
		QueryPrefix, // {what_is}
		ConceptNote,
		LanguageComponent{Category: "Filler", Value: "do i have"},
		TerminatorPunctuation,
	))

	// rule.note.read1
	g.RegisterRule(lib.BuildRule("rule.note.read1", "manage_notes", 0.99,
		LanguageComponent{Category: "ActionPrefix", Value: "read|open|show"},
		LanguageComponent{Category: "Connector", Value: "my"},
		SlotTitle,
		ConceptNote,
		TerminatorPunctuation,
	))

	// rule.note.read2
	g.RegisterRule(lib.BuildRule("rule.note.read2", "manage_notes", 0.99,
		LanguageComponent{Category: "ActionPrefix", Value: "show"},
		LanguageComponent{Category: "Connector", Value: "my"},
		ConceptNote,
		LanguageComponent{Category: "RequiredConnector", Value: "called|titled"},
		SlotTitle,
		TerminatorPunctuation,
	))

	// rule.note.delete1
	g.RegisterRule(lib.BuildRule("rule.note.delete1", "manage_notes", 0.99,
		DeleteActionPrefix,
		LanguageComponent{Category: "Connector", Value: "my"},
		SlotTitle,
		ConceptNote,
		TerminatorPunctuation,
	))

	// rule.note.delete2
	g.RegisterRule(lib.BuildRule("rule.note.delete2", "manage_notes", 0.99,
		DeleteActionPrefix,
		LanguageComponent{Category: "Connector", Value: "the|my"},
		ConceptNote,
		LanguageComponent{Category: "RequiredConnector", Value: "called|titled"},
		SlotTitle,
		TerminatorPunctuation,
	))

	// rule.note.delete3
	g.RegisterRule(lib.BuildRule("rule.note.delete3", "manage_notes", 0.99,
		DeleteActionPrefix,
		ConnectorMy,
		ConceptNote,
		ConnectorCalled,
		SlotTitle,
		TerminatorPunctuation,
	))
}
