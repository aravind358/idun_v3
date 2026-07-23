# Native Network Capability

## Purpose
The Native Network Capability provides direct mechanical access to operating system networking primitives for IDUN V3. Following the Phase 3A Capability Philosophy, it strictly separates the mechanical network transport (TCP, UDP, HTTP, DNS) from any cognitive interpretation, protocol comprehension (like REST/OAuth), or business logic.

## Architecture
- **Router**: Uses `execute.go` and `execute_*.go` handlers mapped to strongly typed enum operations.
- **Provider Interface**: The `NetworkProvider` abstracts standard library `net` and `net/http` dependencies.
- **Implementations**: `NativeProvider` executes the actual networking operations over the host's network interfaces, while `MockProvider` enables completely isolated, stateless testing.

## Operations
- **DNS**: `resolve_dns`, `lookup_ip`
- **HTTP**: `http_get`, `http_post`, `http_head`
- **Sockets**: `open_tcp_socket`, `open_udp_socket`, `close_socket`
- **Data Transfer**: `download`, `upload`
- **Diagnostics**: `list_interfaces`, `get_interface`, `connection_status`, `ping`

## Safety
Capability actively rejects invalid ports, timeout values, and malformed network inputs. Permission checks gate access independently for different classes of network operations.
