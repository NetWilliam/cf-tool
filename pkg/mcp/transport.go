package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Transport represents an MCP transport layer
type Transport interface {
	Send(ctx context.Context, msg *JSONRPCMessage) error
	Receive(ctx context.Context) (*JSONRPCMessage, error)
	Close() error
}

// SSETransport implements Server-Sent Events transport for MCP
type SSETransport struct {
	client      *http.Client
	serverURL   string
	eventChan   chan *JSONRPCMessage
	mu          sync.Mutex
	connected   bool
	stopChan    chan struct{}
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(serverURL string) (*SSETransport, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("server URL cannot be empty")
	}

	return &SSETransport{
		client:    &http.Client{},
		serverURL: serverURL,
		eventChan: make(chan *JSONRPCMessage, 100),
		stopChan:  make(chan struct{}),
		connected: false,
	}, nil
}

// connect establishes SSE connection and starts listening for events
func (t *SSETransport) connect(ctx context.Context) error {
	t.mu.Lock()
	if t.connected {
		t.mu.Unlock()
		return nil
	}
	t.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, "GET", t.serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	t.mu.Lock()
	t.connected = true
	t.mu.Unlock()

	// Start reading SSE events in background
	go t.readEvents(resp.Body)

	return nil
}

// readEvents reads SSE events from the response body
func (t *SSETransport) readEvents(body io.ReadCloser) {
	defer body.Close()
	scanner := bufio.NewScanner(body)

	for {
		select {
		case <-t.stopChan:
			return
		default:
		}

		if !scanner.Scan() {
			return
		}

		line := scanner.Text()

		// SSE format: "data: {...}"
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			var msg JSONRPCMessage
			if err := json.Unmarshal([]byte(data), &msg); err == nil {
				select {
				case t.eventChan <- &msg:
				case <-t.stopChan:
					return
				}
			}
		}
	}
}

// Send sends a JSON-RPC message via SSE (using POST for requests)
func (t *SSETransport) Send(ctx context.Context, msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// For SSE streamable-http, we send requests via POST to the same endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", t.serverURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// Receive receives a JSON-RPC message from SSE event stream
func (t *SSETransport) Receive(ctx context.Context) (*JSONRPCMessage, error) {
	// For streamable-http, responses come directly in the POST response body
	// This is a no-op as Send() already gets the response
	return nil, fmt.Errorf("use SendReceive for SSE transport")
}

// Close closes the SSE transport
func (t *SSETransport) Close() error {
	close(t.stopChan)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connected = false
	return nil
}

// SendReceive sends a message and receives response (for SSE transport)
func (t *SSETransport) SendReceive(ctx context.Context, msg *JSONRPCMessage) (*JSONRPCMessage, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send request via POST
	req, err := http.NewRequestWithContext(ctx, "POST", t.serverURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var response JSONRPCMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// HTTPTransport implements HTTP-based transport for MCP
type HTTPTransport struct {
	client    *http.Client
	serverURL string
	mu        sync.Mutex
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(serverURL string) (*HTTPTransport, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("server URL cannot be empty")
	}

	return &HTTPTransport{
		client:    &http.Client{Timeout: 30 * time.Second},
		serverURL: serverURL,
	}, nil
}

// Send sends a JSON-RPC message via HTTP
func (t *HTTPTransport) Send(ctx context.Context, msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.serverURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// Receive receives a JSON-RPC message via HTTP
func (t *HTTPTransport) Receive(ctx context.Context) (*JSONRPCMessage, error) {
	// For HTTP transport, we send and receive in the same request (Send method)
	// This is a no-op for HTTP streaming transport
	return nil, fmt.Errorf("HTTP transport uses send/receive in one call")
}

// SendReceive sends a message and waits for response via HTTP
func (t *HTTPTransport) SendReceive(ctx context.Context, msg *JSONRPCMessage) (*JSONRPCMessage, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.serverURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var response JSONRPCMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// Close closes the HTTP transport
func (t *HTTPTransport) Close() error {
	// No resources to clean up for HTTP transport
	return nil
}

// StdioTransport implements stdio-based transport
type StdioTransport struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner
	mu      sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(command string, args []string) (*StdioTransport, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	return &StdioTransport{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		scanner: bufio.NewScanner(stdout),
	}, nil
}

// Send sends a JSON-RPC message
func (t *StdioTransport) Send(ctx context.Context, msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// MCP stdio protocol requires messages to be JSON lines
	if _, err := fmt.Fprintln(t.stdin, string(data)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// Receive receives a JSON-RPC message
func (t *StdioTransport) Receive(ctx context.Context) (*JSONRPCMessage, error) {
	done := make(chan error, 1)
	var msg JSONRPCMessage

	go func() {
		if t.scanner.Scan() {
			line := t.scanner.Text()
			done <- json.Unmarshal([]byte(line), &msg)
		} else {
			done <- t.scanner.Err()
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("failed to receive message: %w", err)
		}
		return &msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("receive timeout")
	}
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var errs []error

	if t.stdin != nil {
		if err := t.stdin.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if t.stdout != nil {
		if err := t.stdout.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if t.cmd != nil && t.cmd.Process != nil {
		if err := t.cmd.Process.Kill(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}
