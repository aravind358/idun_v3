package understanding

// LanguageComponent represents a semantic, deterministic block of English language.
// It is used by the Grammar Compiler to deterministically generate regex at startup.
type LanguageComponent struct {
	ID       string // e.g., "LANG_SYS_BATTERY"
	Name     string // e.g., "Battery"
	Value    string // e.g., "battery" (Raw text, NO regex)
	Category string // e.g., "SystemConcept"

	// Reserved for future:
	// Aliases []string
	// Metadata map[string]string
}
