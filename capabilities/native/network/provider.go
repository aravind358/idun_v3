package network

import "context"

// NetworkProvider abstracts native networking operations.
// This isolates the capability from host APIs and enables seamless mock testing.
type NetworkProvider interface {
	// DNS
	ResolveDNS(ctx context.Context, hostname string) ([]string, error)
	LookupIP(ctx context.Context, ip string) ([]string, error)

	// HTTP
	HTTPRequest(ctx context.Context, method, url string, headers map[string]string, body []byte, timeoutMs int) (map[string]interface{}, error)
	Download(ctx context.Context, url, destination string, timeoutMs int) error
	Upload(ctx context.Context, url, source string, timeoutMs int) error

	// Sockets
	OpenTCPSocket(ctx context.Context, address string, timeoutMs int) (string, error)
	OpenUDPSocket(ctx context.Context, address string, timeoutMs int) (string, error)
	CloseSocket(ctx context.Context, socketID string) error

	// Interfaces
	ListInterfaces(ctx context.Context) ([]map[string]interface{}, error)
	GetInterface(ctx context.Context, name string) (map[string]interface{}, error)

	// Status
	ConnectionStatus(ctx context.Context) (map[string]interface{}, error)
	Ping(ctx context.Context, address string, timeoutMs int) (map[string]interface{}, error)
}
