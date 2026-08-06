package understanding

// DefaultTypos maps common misspellings to their correct canonical forms.
var DefaultTypos = map[string]string{
	"waht":   "what",
	"tiem":   "time",
	"wether": "weather",
	"cencel": "cancel",
	"statas": "status",
	"sttaus": "status",
	"helo":   "hello",
	"alrm":   "alarm",
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
