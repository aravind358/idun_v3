package understanding

var (
	// QueryPrefix captures standard information retrieval intents.
	QueryPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_QUERY",
		Name:     "Query Prefix",
		Value:    "what_is|check|query|show|tell me|display|how much",
		Category: "Prefix",
	}

	// CreateActionPrefix captures intents to create or set something.
	CreateActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_CREATE",
		Name:     "Create Action Prefix",
		Value:    "create|make|set",
		Category: "ActionPrefix",
	}

	// DeleteActionPrefix captures intents to delete or remove something.
	DeleteActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_DELETE",
		Name:     "Delete Action Prefix",
		Value:    "delete|remove|erase",
		Category: "ActionPrefix",
	}

	// MoveActionPrefix captures intents to move or transfer something.
	MoveActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_MOVE",
		Name:     "Move Action Prefix",
		Value:    "move|transfer",
		Category: "ActionPrefix",
	}

	// CopyActionPrefix captures intents to copy something.
	CopyActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_COPY",
		Name:     "Copy Action Prefix",
		Value:    "copy",
		Category: "ActionPrefix",
	}

	// RenameActionPrefix captures intents to rename something.
	RenameActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_RENAME",
		Name:     "Rename Action Prefix",
		Value:    "rename",
		Category: "ActionPrefix",
	}

	// OpenActionPrefix captures intents to open or read something.
	OpenActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_OPEN",
		Name:     "Open Action Prefix",
		Value:    "open|read",
		Category: "ActionPrefix",
	}

	// ListActionPrefix captures intents to list or show contents.
	ListActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_LIST",
		Name:     "List Action Prefix",
		Value:    "list|show|read",
		Category: "ActionPrefix",
	}

	// WriteActionPrefix captures intents to write to files.
	WriteActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_WRITE",
		Name:     "Write Action Prefix",
		Value:    "write|append",
		Category: "ActionPrefix",
	}

	// ExistsActionPrefix captures intents to check existence.
	ExistsActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_EXISTS",
		Name:     "Exists Action Prefix",
		Value:    "check if|does",
		Category: "ActionPrefix",
	}

	// DangerousActionPrefix captures dangerous operations.
	DangerousActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_DANGEROUS",
		Name:     "Dangerous Action Prefix",
		Value:    "format drive|erase disk",
		Category: "ActionPrefix",
	}

	// MkdirActionPrefix specifically for mkdir.
	MkdirActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_MKDIR",
		Name:     "Mkdir Action Prefix",
		Value:    "mkdir",
		Category: "ActionPrefix",
	}

	// RemindActionPrefix captures intents to remind.
	RemindActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_REMIND",
		Name:     "Remind Action Prefix",
		Value:    "remind",
		Category: "ActionPrefix",
	}

	// TakeActionPrefix captures intents to take a note.
	TakeActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_TAKE",
		Name:     "Take Action Prefix",
		Value:    "take",
		Category: "ActionPrefix",
	}

	// NoteActionPrefix captures intents starting with 'note'.
	NoteActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_NOTE",
		Name:     "Note Action Prefix",
		Value:    "note",
		Category: "ActionPrefix",
	}

	// SaveActionPrefix captures intents starting with 'save'.
	SaveActionPrefix = LanguageComponent{
		ID:       "LANG_PREFIX_SAVE",
		Name:     "Save Action Prefix",
		Value:    "save",
		Category: "ActionPrefix",
	}
)
