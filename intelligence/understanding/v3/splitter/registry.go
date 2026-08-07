package splitter

import (
	"sort"
	"strings"
)

// ConnectorRegistry manages the list of connectors used for utterance segmentation.
type ConnectorRegistry struct {
	connectors []string
}

// NewConnectorRegistry creates a new registry with default connectors.
func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		connectors: []string{
			" and then ",
			" after that ",
			" as well as ",
			" and ",
			" then ",
			" also ",
			", and ",
			", then ",
			", ",
			" plus ",
			" besides ",
			" afterwards ",
			" along with ",
		},
	}
}

// AddConnector adds a new connector to the registry.
func (r *ConnectorRegistry) AddConnector(c string) {
	r.connectors = append(r.connectors, c)
}

// GetConnectors returns the list of connectors sorted by length descending,
// which is required to prevent partial matches (e.g., matching " and " before " and then ").
func (r *ConnectorRegistry) GetConnectors() []string {
	sorted := make([]string, len(r.connectors))
	copy(sorted, r.connectors)
	
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i]) > len(sorted[j])
	})
	
	return sorted
}

// ReplaceConnectorsWithDelim replaces all registered connectors in the given string
// with a canonical delimiter, enabling straightforward O(N) splitting.
func (r *ConnectorRegistry) ReplaceConnectorsWithDelim(utterance string, delim string) string {
	sorted := r.GetConnectors()
	result := utterance
	for _, c := range sorted {
		// Replace case-insensitively. Since strings.ReplaceAll is case-sensitive,
		// we must do this manually or assume connectors are lowercase and we lower the string.
		// For robustness in this implementation, we can use a naive approach or regex for case-insensitive replacement.
		// For simplicity, we'll do case-insensitive replacement manually.
		result = replaceAllCaseInsensitive(result, c, delim)
	}
	return result
}

// replaceAllCaseInsensitive replaces all instances of search (case-insensitive) with replace.
func replaceAllCaseInsensitive(str, search, replace string) string {
	lowerStr := strings.ToLower(str)
	lowerSearch := strings.ToLower(search)
	
	var sb strings.Builder
	lastMatch := 0
	
	for {
		idx := strings.Index(lowerStr[lastMatch:], lowerSearch)
		if idx == -1 {
			sb.WriteString(str[lastMatch:])
			break
		}
		
		matchStart := lastMatch + idx
		sb.WriteString(str[lastMatch:matchStart])
		sb.WriteString(replace)
		
		lastMatch = matchStart + len(search)
	}
	
	return sb.String()
}
