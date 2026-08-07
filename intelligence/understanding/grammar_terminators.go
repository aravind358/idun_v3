package understanding

var (
	TerminatorPunctuation = LanguageComponent{
		ID:       "LANG_TERM_PUNCTUATION",
		Name:     "Punctuation Terminator",
		Value:    "punctuation",
		Category: "Terminator",
	}

	ActionSuffixExists = LanguageComponent{
		ID:       "LANG_SUFFIX_EXISTS",
		Name:     "Exists Suffix",
		Value:    "exist|exists",
		Category: "Suffix",
	}

	FallbackMatchAll = LanguageComponent{
		ID:       "LANG_SYS_FALLBACK",
		Name:     "Fallback Match All",
		Value:    ".*",
		Category: "Terminator", // Treat as terminator so it skips standard wrapping
	}
)
