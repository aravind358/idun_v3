package network

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type NativeProvider struct {
	mu      sync.Mutex
	sockets map[string]net.Conn
	counter int
}

func NewNativeProvider() *NativeProvider {
	return &NativeProvider{
		sockets: make(map[string]net.Conn),
	}
}

func (p *NativeProvider) ResolveDNS(ctx context.Context, hostname string) ([]string, error) {
	return net.DefaultResolver.LookupHost(ctx, hostname)
}

func (p *NativeProvider) LookupIP(ctx context.Context, ip string) ([]string, error) {
	return net.DefaultResolver.LookupAddr(ctx, ip)
}

func (p *NativeProvider) HTTPRequest(ctx context.Context, method, url string, headers map[string]string, body []byte, timeoutMs int) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     respHeaders,
		"body":        respBody,
	}, nil
}

func (p *NativeProvider) Download(ctx context.Context, url, destination string, timeoutMs int) error {
	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("download failed with status: " + resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (p *NativeProvider) Upload(ctx context.Context, url, source string, timeoutMs int) error {
	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, f)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("upload failed with status: " + resp.Status)
	}
	return nil
}

func (p *NativeProvider) OpenTCPSocket(ctx context.Context, address string, timeoutMs int) (string, error) {
	d := net.Dialer{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return "", err
	}
	return p.registerSocket(conn), nil
}

func (p *NativeProvider) OpenUDPSocket(ctx context.Context, address string, timeoutMs int) (string, error) {
	d := net.Dialer{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	conn, err := d.DialContext(ctx, "udp", address)
	if err != nil {
		return "", err
	}
	return p.registerSocket(conn), nil
}

func (p *NativeProvider) CloseSocket(ctx context.Context, socketID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	conn, ok := p.sockets[socketID]
	if !ok {
		return errors.New("invalid socket id")
	}
	delete(p.sockets, socketID)
	return conn.Close()
}

func (p *NativeProvider) registerSocket(conn net.Conn) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.counter++
	id := "sock-" + string(rune(p.counter)) // basic id
	p.sockets[id] = conn
	return id
}

func (p *NativeProvider) ListInterfaces(ctx context.Context) ([]map[string]interface{}, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var results []map[string]interface{}
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		var addrStrs []string
		for _, a := range addrs {
			addrStrs = append(addrStrs, a.String())
		}
		results = append(results, map[string]interface{}{
			"index":        i.Index,
			"name":         i.Name,
			"mac_address":  i.HardwareAddr.String(),
			"addresses":    addrStrs,
			"flags":        i.Flags.String(),
		})
	}
	return results, nil
}

func (p *NativeProvider) GetInterface(ctx context.Context, name string) (map[string]interface{}, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}
	addrs, _ := iface.Addrs()
	var addrStrs []string
	for _, a := range addrs {
		addrStrs = append(addrStrs, a.String())
	}
	return map[string]interface{}{
		"index":        iface.Index,
		"name":         iface.Name,
		"mac_address":  iface.HardwareAddr.String(),
		"addresses":    addrStrs,
		"flags":        iface.Flags.String(),
	}, nil
}

func (p *NativeProvider) ConnectionStatus(ctx context.Context) (map[string]interface{}, error) {
	// A simple mechanical ping to check if network is reachable.
	_, err := net.LookupHost("google.com")
	connected := err == nil
	return map[string]interface{}{
		"connected": connected,
	}, nil
}

func (p *NativeProvider) Ping(ctx context.Context, address string, timeoutMs int) (map[string]interface{}, error) {
	// True ICMP ping requires root on most OSes. We do a TCP ping to port 80/443 as a fallback.
	// For mechanical native transport, an actual ICMP ping library would be used or exec.Command("ping").
	// Using a basic TCP dial to port 80 as a synthetic ping for now.
	start := time.Now()
	d := net.Dialer{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", address+":80")
	if err != nil {
		return map[string]interface{}{"reachable": false}, nil
	}
	conn.Close()
	return map[string]interface{}{
		"reachable": true,
		"latency":   time.Since(start).Milliseconds(),
	}, nil
}
