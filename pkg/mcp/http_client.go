package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPClient implements simple HTTP-based MCP client (streamable-http)
type HTTPClient struct {
	client     *http.Client
	serverURL  string
	sessionID  string
	mu         sync.RWMutex
	initialized bool
	requestID   int64
}

// NewHTTPClient creates a new HTTP MCP client
func NewHTTPClient(serverURL string) (*HTTPClient, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("server URL cannot be empty")
	}

	return &HTTPClient{
		client:      &http.Client{Timeout: 30 * time.Second},
		serverURL:  serverURL,
		initialized: false,
		requestID:   0,
	}, nil
}

// Initialize initializes the MCP session
func (c *HTTPClient) Initialize(ctx context.Context) error {
	req := &InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{},
		ClientInfo: map[string]string{
			"name":    "cf-tool",
			"version": "1.0.0",
		},
	}

	msg := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextID(),
		Method:  "initialize",
	}

	params, _ := json.Marshal(req)
	msg.Params = json.RawMessage(params)

	resp, err := c.sendRequest(ctx, msg, nil)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.initialized = true
	return nil
}

// CallTool calls an MCP tool
func (c *HTTPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not initialized")
	}
	sessionID := c.sessionID
	c.mu.RUnlock()

	req := &CallToolRequest{
		Name:      name,
		Arguments: args,
	}

	msg := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextID(),
		Method:  "tools/call",
	}

	params, _ := json.Marshal(req)
	msg.Params = json.RawMessage(params)

	// Use session ID if available
	var err error
	var resp *JSONRPCMessage
	if sessionID != "" {
		resp, err = c.sendRequest(ctx, msg, &sessionID)
	} else {
		resp, err = c.sendRequest(ctx, msg, nil)
	}

	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool call error: %s", resp.Error.Message)
	}

	var result ToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

// ListTools lists available tools
func (c *HTTPClient) ListTools(ctx context.Context) ([]Tool, error) {
	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not initialized")
	}
	sessionID := c.sessionID
	c.mu.RUnlock()

	msg := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextID(),
		Method:  "tools/list",
	}

	// Use session ID if available
	var err error
	var resp *JSONRPCMessage
	if sessionID != "" {
		resp, err = c.sendRequest(ctx, msg, &sessionID)
	} else {
		resp, err = c.sendRequest(ctx, msg, nil)
	}

	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", resp.Error.Message)
	}

	var listResp ListToolsResponse
	if err := json.Unmarshal(resp.Result, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return listResp.Tools, nil
}

// Ping tests the connection
func (c *HTTPClient) Ping(ctx context.Context) error {
	_, err := c.ListTools(ctx)
	return err
}

// Close closes the client
func (c *HTTPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.initialized = false
	return nil
}

// Call sends a JSON-RPC request and returns the response
func (c *HTTPClient) Call(ctx context.Context, msg *JSONRPCMessage) (*JSONRPCMessage, error) {
	c.mu.RLock()
	sessionID := c.sessionID
	c.mu.RUnlock()

	// For first request after initialize, sessionID might be empty
	// sendRequest will handle this
	if sessionID == "" {
		return c.sendRequest(ctx, msg, nil)
	}
	return c.sendRequest(ctx, msg, &sessionID)
}

// sendRequest sends a JSON-RPC request via HTTP POST
func (c *HTTPClient) sendRequest(ctx context.Context, msg *JSONRPCMessage, sessionID *string) (*JSONRPCMessage, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	// Add session ID header if we have one
	if sessionID != nil && *sessionID != "" {
		req.Header.Set("mcp-session-id", *sessionID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse SSE response
	var response JSONRPCMessage
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse SSE format from body
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		// SSE format: "data: {...}"
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			// Try to parse the JSON
			if err := json.Unmarshal([]byte(data), &response); err == nil {
				// Successfully parsed, break out of loop
				break
			}
		}
	}

	// Extract session ID from response headers if present
	if sessionIDHeader := resp.Header.Get("mcp-session-id"); sessionIDHeader != "" {
		c.mu.Lock()
		c.sessionID = sessionIDHeader
		c.mu.Unlock()
	}

	return &response, nil
}

// nextID generates the next request ID
func (c *HTTPClient) nextID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestID++
	return c.requestID
}
