package understanding

// DefaultTypos maps common misspellings to their correct canonical forms.
var DefaultTypos = map[string]string{
	// General/Misc
	"waht":      "what",
	"wut":       "what",
	"wat":       "what",
	"cencel":    "cancel",
	"cancle":    "cancel",
	"statas":    "status",
	"sttaus":    "status",
	"alrm":      "alarm",
	
	// Greetings
	"helo":      "hello",
	"hlo":       "hello",
	"hallo":     "hello",
	"heyy":      "hey",
	
	// Time & Date
	"tiem":      "time",
	"tim":       "time",
	"dat":       "date",
	"daet":      "date",
	"toady":     "today",
	"tomorow":   "tomorrow",
	"tommorow":  "tomorrow",
	"tommorrow": "tomorrow",
	
	// Weather
	"wether":    "weather",
	"waether":   "weather",
	"wheather":  "weather",
	"weatar":    "weather",
	
	// Battery & System
	"battrey":   "battery",
	"battry":    "battery",
	"battary":   "battery",
	"batery":    "battery",
	"memry":     "memory",
	"memary":    "memory",
	"memeory":   "memory",
	"shutdwn":   "shutdown",
	"shtdown":   "shutdown",
	
	// Files & Notes
	"not":       "note",
	"noet":      "note",
	"crete":     "create",
	"craete":    "create",
	"foldar":    "folder",
	"floder":    "folder",
	"fiel":      "file",
	"delte":     "delete",
	"delet":     "delete",
	"remov":     "remove",
	"renam":     "rename",
	
	// Calculator
	"calclator": "calculator",
	"calulator": "calculator",
	"calcualtor": "calculator",
	"calculater": "calculator",
	
	// Locations
	"hyderbad":   "hyderabad",
	"heyderabad": "hyderabad",
	"hyderabd":   "hyderabad",
	"seatle":     "seattle",
	"seatel":     "seattle",
	
	// Help
	"halp":      "help",
	"hlp":       "help",
}

// DefaultFillers lists conversational filler phrases that should be stripped
// during normalization to simplify grammar matching, while remaining recorded
// in the NormalizationProfile. These should be ordered by length descending if processed sequentially.
var DefaultFillers = []string{
	"i would like to know",
	"would you be able to",
	"i'd like to know",
	"i was wondering",
	"i want to know",
	"please show me",
	"please tell me",
	"would you mind",
	"could i know",
	"if possible",
	"let me know",
	"may i know",
	"real quick",
	"basically",
	"could you",
	"would you",
	"actually",
	"can you",
	"quickly",
	"show me",
	"tell me",
	"for me",
	"kindly",
	"please",
	"simply",
	"can u",
	"just",
	"pls",
	"plz",
}

// DefaultSynonyms maps canonical words/phrases to a list of equivalent synonyms.
// The normalizer replaces occurrences of each synonym in the cleaned input
// with the canonical key before grammar matching.
var DefaultSynonyms = map[string][]string{
	// --- Existing groups (preserved) ---
	"status":   {"system status", "current status"},
	"cancel":   {"stop", "abort", "nevermind", "never mind", "quit", "exit", "terminate", "halt"},
	"hello":    {"hi", "hey", "greetings", "good morning", "good afternoon", "good evening", "hi there", "howdy", "morning", "what's up"},
	"goodbye":  {"bye", "see you", "farewell", "cya"},
	"what_is":  {"what is", "what's", "how is", "how's", "what"},
	"set":      {"create", "make", "add"},
	"alarm":    {"timer", "alert"},

	// --- Battery (U2.1) ---
	// Grammar rule.sys.battery1 matches: battery[level|percentage|status]
	// So we normalize all charge/power phrases → "battery"
	"battery": {
		// --- U2.1 ---
		"battery status",
		"battery level",
		"battery percentage",
		"battery percent",
		"battery remaining",
		"battery left",
		"charge level",
		"charge status",
		"power level",
		"remaining battery",
		// --- U2.2 ---
		"how much battery do i have",
		"what is my battery level",
		"what is my battery percentage",
		"is my laptop charging",
		"is the battery charging",
		"how much power is left",
		"check the battery",
		"how much battery is left",
		"is my pc charging",
		"is my computer charging",
		"is pc charging",
		"is computer charging",
		"how much power do i have",
	},

	// --- Weather (U2.1) ---
	// Existing {weather} placeholder already covers "temperature","forecast","conditions"
	// Add further weather phrases that normalize to "weather"
	"weather": {
		// --- U2.1 ---
		"temperature",
		"forecast",
		"conditions",
		"outside weather",
		"weather report",
		"weather status",
		"outside temperature",
		"current weather",
		// --- U2.2 ---
		"check the weather",
		"weather outside",
		"what is it like outside",
		"what are the conditions outside",
		"what is it like outside today",
		"weather today",
		"it like outside",
	},

	// --- Time (U2.1) ---
	// Grammar rule.time matches: [what is] [the|current] time [is it|now]
	// "clock", "current time", "local time", "time now" must normalize to "time"
	"time": {
		// --- U2.1 ---
		"clock",
		"current time",
		"local time",
		"time now",
		"present time",
		// --- U2.2 ---
		"what time is it",
		"tell me the time",
		"what is the current time",
		"do you have the time",
		"do you have the clock",
		"do you have time",
		"what does the time say",
		"what does the clock say",
		"does the time say",
		"does the clock say",
		"time is it",
	},

	// --- Date (U2.1) ---
	// Grammar rule.date matches: [what is] [today's|current|the] date [is today]
	// Note: "today" is NOT added here because the grammar handles "today's date" natively;
	// replacing "today" inside "today's date" would produce "date's date".
	"date": {
		// --- U2.1 ---
		"current date",
		"calendar date",
		"present date",
		// --- U2.2 ---
		"what is today's date",
		"what date is it",
		"tell me today's date",
		"today's date",
		"today's day",
		"what day is it",
		"date is today",
		"day is today",
		"date is it",
		"day is it",
		"the date today",
	},

	// --- Memory (U2.1) ---
	// Grammar rule.sys.mem1 matches: [what is|check|query] memory [usage]
	// "ram", "ram usage", "memory usage", "available memory" → "memory"
	"memory": {
		// --- U2.1 ---
		"ram",
		"ram usage",
		"memory usage",
		"memory status",
		"available memory",
		"used memory",
		// --- U2.2 ---
		"how much memory is being used",
		"how much ram is available",
		"check memory usage",
		"show ram usage",
		"available ram",
		"how much memory is available",
		"how much memory is used",
		"how much ram is used",
		"system memory",
		"system ram",
		"is there available memory",
		"is there memory",
		"how much system memory is being used",
	},

	// --- Help (U2.1) ---
	// Grammar rule.system.help matches: help|what can you do|how to use
	// "assist", "support", "guide", "instructions", "manual" → "help"
	"help": {
		// --- U2.1 ---
		"assist",
		"support",
		"guide",
		"instructions",
		"manual",
		// --- U2.2 ---
		"i need help",
		"help me",
		"the help",
		"open help",
		"i need assistance",
		"show instructions",
		"what can you do",
		"give me some tips",
	},
}

// DefaultEntities maps a canonical entity representation to its common aliases or variations.
// It is used to normalize domain entities (locations, applications, devices, file types) 
// to exactly match their required canonical form for grammar and execution.
var DefaultEntities = map[string][]string{
	// Cities
	"Hyderabad": {"hyderabad", "hyderbad", "hyderabd", "heyderabad", "hydrabad"},
	"Mumbai":    {"mumbai", "bombay"},
	"Bengaluru": {"bangalore", "bengaluru"},
	
	// Applications
	"VS Code": {"vscode", "vs code", "visual studio code"},
	"Chrome":  {"chrome browser", "google chrome", "chrome"},
	"Notepad": {"notepad.exe", "notepad"},
	
	// Devices / Operating System
	"computer": {"computer", "machine", "desktop", "pc"},
	"laptop":   {"notebook", "laptop"},
	
	// File Types
	"image":     {"picture", "photo", "image", "pic"},
	"document":  {"document", "doc"},
	"directory": {"directory", "folder"},
}


// DefaultContractions maps common English contractions to their expanded forms.
var DefaultContractions = map[string]string{
	"what's":    "what is",
	"it's":      "it is",
	"i'm":       "i am",
	"i'd":       "i would",
	"i'll":      "i will",
	"i've":      "i have",
	"don't":     "do not",
	"doesn't":   "does not",
	"didn't":    "did not",
	"can't":     "cannot",
	"couldn't":  "could not",
	"wouldn't":  "would not",
	"shouldn't": "should not",
	"won't":     "will not",
	"you're":    "you are",
	"we're":     "we are",
	"they're":   "they are",
	"there's":   "there is",
	"that's":    "that is",
	"who's":     "who is",
	"where's":   "where is",
	"when's":    "when is",
	"why's":     "why is",
	"how's":     "how is",
}
