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
	"would you mind",
	"could you",
	"would you",
	"can you",
	"please",
	"kindly",
	"just",
	"for me",
	"tell me",
	"show me",
}

// DefaultSynonyms maps canonical actions to a list of equivalent words/phrases.
var DefaultSynonyms = map[string][]string{
	"weather":  {"temperature", "forecast", "conditions"},
	"status":   {"system status", "current status"},
	"cancel":   {"stop", "abort", "nevermind", "never mind"},
	"hello":    {"hi", "hey", "greetings", "good morning", "good afternoon", "good evening"},
	"goodbye":  {"bye", "see you", "farewell", "cya"},
	"what_is":  {"what is", "what's", "how is", "how's", "what"},
	"set":      {"create", "make", "add"},
	"alarm":    {"timer", "alert"},
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
