package network

// NetworkOperation defines strongly-typed constants for permitted operations.
type NetworkOperation string

const (
	OperationResolveDNS       NetworkOperation = "resolve_dns"
	OperationLookupIP         NetworkOperation = "lookup_ip"
	OperationHTTPGet          NetworkOperation = "http_get"
	OperationHTTPPost         NetworkOperation = "http_post"
	OperationHTTPHead         NetworkOperation = "http_head"
	OperationDownload         NetworkOperation = "download"
	OperationUpload           NetworkOperation = "upload"
	OperationOpenTCPSocket    NetworkOperation = "open_tcp_socket"
	OperationOpenUDPSocket    NetworkOperation = "open_udp_socket"
	OperationCloseSocket      NetworkOperation = "close_socket"
	OperationListInterfaces   NetworkOperation = "list_interfaces"
	OperationGetInterface     NetworkOperation = "get_interface"
	OperationConnectionStatus NetworkOperation = "connection_status"
	OperationPing             NetworkOperation = "ping"
)

// IsValid validates if a string matches a known NetworkOperation.
func (o NetworkOperation) IsValid() bool {
	switch o {
	case OperationResolveDNS, OperationLookupIP, OperationHTTPGet,
		OperationHTTPPost, OperationHTTPHead, OperationDownload,
		OperationUpload, OperationOpenTCPSocket, OperationOpenUDPSocket,
		OperationCloseSocket, OperationListInterfaces, OperationGetInterface,
		OperationConnectionStatus, OperationPing:
		return true
	}
	return false
}
