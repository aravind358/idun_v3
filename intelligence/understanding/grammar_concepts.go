package understanding

var (
	// System Concepts
	ConceptBattery = LanguageComponent{
		ID:       "LANG_SYS_BATTERY",
		Name:     "Battery",
		Value:    "battery",
		Category: "SystemConcept",
	}

	ConceptCPU = LanguageComponent{
		ID:       "LANG_SYS_CPU",
		Name:     "CPU",
		Value:    "cpu",
		Category: "SystemConcept",
	}

	ConceptMemory = LanguageComponent{
		ID:       "LANG_SYS_MEMORY",
		Name:     "Memory",
		Value:    "memory",
		Category: "SystemConcept",
	}

	ConceptDisk = LanguageComponent{
		ID:       "LANG_SYS_DISK",
		Name:     "Disk",
		Value:    "disk",
		Category: "SystemConcept",
	}

	// Environment Concepts
	ConceptWeather = LanguageComponent{
		ID:       "LANG_ENV_WEATHER",
		Name:     "Weather",
		Value:    "weather",
		Category: "EnvironmentConcept",
	}

	// Time Concepts
	ConceptTime = LanguageComponent{
		ID:       "LANG_TIME_TIME",
		Name:     "Time",
		Value:    "time",
		Category: "TimeConcept",
	}

	ConceptDate = LanguageComponent{
		ID:       "LANG_TIME_DATE",
		Name:     "Date",
		Value:    "date|day",
		Category: "TimeConcept",
	}

	ConceptAlarm = LanguageComponent{
		ID:       "LANG_TIME_ALARM",
		Name:     "Alarm",
		Value:    "alarm",
		Category: "TimeConcept",
	}

	// File Concepts
	OptionalConceptFileOrDir = LanguageComponent{
		ID:       "LANG_FILE_OPT_FILEDIR",
		Name:     "Optional File or Directory",
		Value:    "file|folder|directory",
		Category: "OptionalConcept",
	}

	ConceptFilesList = LanguageComponent{
		ID:       "LANG_FILE_FILES_LIST",
		Name:     "Files (list)",
		Value:    "all files|files|contents",
		Category: "OptionalConcept",
	}

	// Filesystem Concepts
	ConceptFile = LanguageComponent{
		ID:       "LANG_FS_FILE",
		Name:     "File",
		Value:    "file|document",
		Category: "FilesystemConcept",
	}

	ConceptDirectory = LanguageComponent{
		ID:       "LANG_FS_DIRECTORY",
		Name:     "Directory",
		Value:    "directory|folder",
		Category: "FilesystemConcept",
	}
	
	ConceptFileOrDir = LanguageComponent{
		ID:       "LANG_FS_FILEORDIR",
		Name:     "File or Directory",
		Value:    "file|folder|directory",
		Category: "FilesystemConcept",
	}

	// Knowledge Concepts
	ConceptNote = LanguageComponent{
		ID:       "LANG_KNOWLEDGE_NOTE",
		Name:     "Note",
		Value:    "note|notes",
		Category: "KnowledgeConcept",
	}

	ConceptReminder = LanguageComponent{
		ID:       "LANG_KNOWLEDGE_REMINDER",
		Name:     "Reminder",
		Value:    "reminder",
		Category: "KnowledgeConcept",
	}

	// Meta Concepts
	ConceptHelp = LanguageComponent{
		ID:       "LANG_META_HELP",
		Name:     "Help",
		Value:    "help|what can you do|how to use",
		Category: "MetaConcept",
	}

	ConceptStatus = LanguageComponent{
		ID:       "LANG_META_STATUS",
		Name:     "Status",
		Value:    "status",
		Category: "MetaConcept",
	}

	ConceptIdentity = LanguageComponent{
		ID:       "LANG_META_IDENTITY",
		Name:     "Identity",
		Value:    "who|what",
		Category: "MetaConcept",
	}

	ConceptGreeting = LanguageComponent{
		ID:       "LANG_META_GREETING",
		Name:     "Greeting",
		Value:    "hello",
		Category: "MetaConcept",
	}

	ConceptFarewell = LanguageComponent{
		ID:       "LANG_META_FAREWELL",
		Name:     "Farewell",
		Value:    "goodbye",
		Category: "MetaConcept",
	}

	ConceptCancel = LanguageComponent{
		ID:       "LANG_META_CANCEL",
		Name:     "Cancel",
		Value:    "cancel",
		Category: "MetaConcept",
	}

	// Context Concepts
	ConceptAnaphora = LanguageComponent{
		ID:       "LANG_CTX_ANAPHORA",
		Name:     "Anaphora",
		Value:    "it|that|this|those|them",
		Category: "Slot", // Mapped as a slot so its value is extracted for the Context Resolver
	}

	ConceptAffirmation = LanguageComponent{
		ID:       "LANG_CTX_AFFIRMATION",
		Name:     "Affirmation",
		Value:    "yes|yeah|yep|sure|do it|confirm|proceed|ok|okay",
		Category: "MetaConcept",
	}

	ConceptNegation = LanguageComponent{
		ID:       "LANG_CTX_NEGATION",
		Name:     "Negation",
		Value:    "no|nope|nah|nevermind|never mind|cancel|stop",
		Category: "MetaConcept",
	}
)
