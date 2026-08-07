package understanding

var (
	// System Properties
	PropertyStatus = LanguageComponent{
		ID:       "LANG_PROP_STATUS",
		Name:     "Status Property",
		Value:    "status|level|percentage",
		Category: "SystemProperty",
	}

	PropertyUsage = LanguageComponent{
		ID:       "LANG_PROP_USAGE",
		Name:     "Usage Property",
		Value:    "usage",
		Category: "SystemProperty",
	}

	PropertySpace = LanguageComponent{
		ID:       "LANG_PROP_SPACE",
		Name:     "Space Property",
		Value:    "space|usage",
		Category: "SystemProperty",
	}

	// Will add more as IDUN capabilities expand (e.g., Temperature, Speed, etc.)
)
