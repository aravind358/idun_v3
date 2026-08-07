package understanding

// LoadFileRules registers all deterministic file and directory rules.
func LoadFileRules(g *DefaultGrammarSpecialist, lib *GrammarBuilderLibrary) {

	// -------------------------------------------------------------------------
	// File Rules
	// -------------------------------------------------------------------------

	// rule.file.move
	g.RegisterRule(lib.BuildRule("rule.file.move", "file_operation", 0.98,
		MoveActionPrefix,
		OptionalConceptFileOrDir,
		SlotSource,
		ConnectorTo,
		SlotDestination,
		TerminatorPunctuation,
	))

	// rule.file.copy
	g.RegisterRule(lib.BuildRule("rule.file.copy", "file_operation", 0.98,
		CopyActionPrefix,
		OptionalConceptFileOrDir,
		SlotSource,
		ConnectorTo,
		SlotDestination,
		TerminatorPunctuation,
	))

	// rule.file.copy_only
	g.RegisterRule(lib.BuildRule("rule.file.copy_only", "file_operation", 0.97,
		CopyActionPrefix,
		OptionalConceptFileOrDir,
		SlotSource,
		TerminatorPunctuation,
	))

	// rule.file.rename
	g.RegisterRule(lib.BuildRule("rule.file.rename", "file_operation", 0.98,
		RenameActionPrefix,
		OptionalConceptFileOrDir,
		SlotSource,
		ConnectorTo,
		SlotDestination,
		TerminatorPunctuation,
	))

	// rule.file.open
	g.RegisterRule(lib.BuildRule("rule.file.open", "file_operation", 0.98,
		OpenActionPrefix,
		OptionalConceptFileOrDir,
		SlotFilename,
		TerminatorPunctuation,
	))

	// rule.file.delete
	g.RegisterRule(lib.BuildRule("rule.file.delete", "file_operation", 0.98,
		DeleteActionPrefix,
		OptionalConceptFileOrDir,
		SlotFilename,
		TerminatorPunctuation,
	))

	// rule.file.write
	g.RegisterRule(lib.BuildRule("rule.file.write", "file_operation", 0.98,
		WriteActionPrefix,
		SlotDataText,
		ConnectorTo,
		OptionalConceptFileOrDir,
		SlotPath,
		TerminatorPunctuation,
	))

	// rule.file.exists
	g.RegisterRule(lib.BuildRule("rule.file.exists", "file_operation", 0.98,
		ExistsActionPrefix,
		OptionalConceptFileOrDir,
		SlotPath,
		ActionSuffixExists,
		TerminatorPunctuation,
	))

	// rule.file.create
	g.RegisterRule(lib.BuildRule("rule.file.create", "file_operation", 0.98,
		CreateActionPrefix,
		ConceptFileOrDir,
		SlotPath,
		TerminatorPunctuation,
	))

	// rule.file.dangerous
	g.RegisterRule(lib.BuildRule("rule.file.dangerous", "file_operation", 0.99,
		DangerousActionPrefix,
		TerminatorPunctuation, // Not really needed if it's strictly format drive, but helps with whitespace
	))

	// -------------------------------------------------------------------------
	// Directory Rules
	// -------------------------------------------------------------------------

	// rule.dir.create1
	g.RegisterRule(lib.BuildRule("rule.dir.create1", "create_directory", 0.99,
		CreateActionPrefix,
		ConnectorA,
		ConceptDirectory,
		ConnectorCalled,
		SlotDirectory,
		TerminatorPunctuation,
	))

	// rule.dir.create2
	g.RegisterRule(lib.BuildRule("rule.dir.create2", "create_directory", 0.99,
		MkdirActionPrefix,
		SlotDirectory,
		TerminatorPunctuation,
	))

	// rule.dir.list1
	g.RegisterRule(lib.BuildRule("rule.dir.list1", "list_files", 0.99,
		ListActionPrefix,
		ConceptFilesList,
		ConnectorIn,
		SlotDirectory,
		TerminatorPunctuation,
	))

	// rule.dir.list2
	g.RegisterRule(lib.BuildRule("rule.dir.list2", "list_files", 0.98,
		ListActionPrefix,
		ConceptDirectory,
		SlotDirectory,
		TerminatorPunctuation,
	))

	// rule.dir.list3
	g.RegisterRule(lib.BuildRule("rule.dir.list3", "list_files", 0.97,
		QueryPrefix, // {what_is}
		ConnectorIn,
		SlotDirectory,
		TerminatorPunctuation,
	))

	// rule.dir.list_simple
	g.RegisterRule(lib.BuildRule("rule.dir.list_simple", "list_files", 0.97,
		ListActionPrefix,
		SlotDirectory,
		TerminatorPunctuation,
	))
}
