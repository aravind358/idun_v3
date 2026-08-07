package understanding

// LoadSystemRules registers the deterministically compiled System capabilities
// (Battery, CPU, Memory, Disk, Power operations) into the GrammarSpecialist.
func LoadSystemRules(g GrammarSpecialist, lib *GrammarBuilderLibrary) {
	// Battery Rules
	g.RegisterRule(lib.BuildObjectPropertyQuery("rule.sys.battery", "query_battery", 0.98, ConceptBattery, PropertyStatus))
	
	// CPU Rules
	g.RegisterRule(lib.BuildObjectPropertyQuery("rule.sys.cpu", "query_cpu", 0.98, ConceptCPU, PropertyUsage))
	
	// Memory Rules
	g.RegisterRule(lib.BuildObjectPropertyQuery("rule.sys.mem", "query_memory", 0.98, ConceptMemory, PropertyUsage))
	
	// Disk Rules
	g.RegisterRule(lib.BuildObjectPropertyQuery("rule.sys.disk", "query_disk", 0.98, ConceptDisk, PropertySpace))

	// System Control - Shutdown
	g.RegisterRule(lib.BuildRule("rule.sys.shutdown", "system_shutdown", 0.98,
		LanguageComponent{ID: "LANG_ACTION_SHUTDOWN", Name: "Shutdown", Value: "shut down|shutdown|turn off", Category: "Action"},
		ConnectorThe,
		LanguageComponent{ID: "LANG_SYS_COMPUTER", Name: "Computer", Value: "computer|system|pc|machine", Category: "SystemConcept"},
		TerminatorPunctuation,
	))

	// System Control - Restart
	g.RegisterRule(lib.BuildRule("rule.sys.restart", "system_restart", 0.98,
		LanguageComponent{ID: "LANG_ACTION_RESTART", Name: "Restart", Value: "restart|reboot", Category: "Action"},
		ConnectorThe,
		LanguageComponent{ID: "LANG_SYS_COMPUTER", Name: "Computer", Value: "computer|system|pc|machine", Category: "SystemConcept"},
		TerminatorPunctuation,
	))
	
	// System Control - Lock
	g.RegisterRule(lib.BuildRule("rule.sys.lock", "system_lock", 0.98,
		LanguageComponent{ID: "LANG_ACTION_LOCK", Name: "Lock", Value: "lock", Category: "Action"},
		ConnectorThe,
		LanguageComponent{ID: "LANG_SYS_COMPUTER_SCREEN", Name: "Screen", Value: "computer|system|pc|machine|screen", Category: "SystemConcept"},
		TerminatorPunctuation,
	))
	
	// Dangerous Ops
	g.RegisterRule(lib.BuildRule("rule.sys.destroy", "system_shutdown", 0.99,
		LanguageComponent{ID: "LANG_ACTION_DESTROY", Name: "Destroy", Value: "destroy|kill|wipe", Category: "Action"},
		ConnectorThe,
		LanguageComponent{ID: "LANG_SYS_ALL", Name: "All Systems", Value: "computer|system|pc|machine|windows|all processes|disk|my disk", Category: "SystemConcept"},
		TerminatorPunctuation,
	))
}
