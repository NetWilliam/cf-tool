package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Client represents an MCP client (wrapper for backward compatibility)
type Client struct {
	httpClient    *HTTPClient
	stdioClient   *StdioClient
	transportType string // "stdio" or "http"
	initialized   bool
	mu            sync.RWMutex
	requestID     int64
}

// NewClient creates a new MCP client with stdio transport
func NewClient(command string, args []string) (*Client, error) {
	stdioClient, err := NewStdioClient(command, args)
	if err != nil {
		return nil, err
	}

	client := &Client{
		stdioClient:   stdioClient,
		transportType: "stdio",
	}

	// Initialize the connection
	if err := client.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return client, nil
}

// NewClientHTTP creates a new MCP client with HTTP transport
func NewClientHTTP(serverURL string) (*Client, error) {
	httpClient, err := NewHTTPClient(serverURL)
	if err != nil {
		return nil, err
	}

	// Initialize the HTTP client
	if err := httpClient.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	client := &Client{
		httpClient:    httpClient,
		transportType: "http",
		initialized:   true, // Mark as initialized since HTTPClient is initialized
	}

	return client, nil
}

// Initialize initializes the MCP session
func (c *Client) Initialize(ctx context.Context) error {
	// Generate request ID BEFORE acquiring lock
	msgID := c.nextID()

	c.mu.Lock()
	defer c.mu.Unlock()

	req := &InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]interface{}{},
		ClientInfo: map[string]string{
			"name":    "cf-tool",
			"version": "1.0.0",
		},
	}

	msg := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      msgID,
		Method:  "initialize",
	}

	params, _ := json.Marshal(req)
	msg.Params = json.RawMessage(params)

	var resp *JSONRPCMessage
	var err error

	if c.transportType == "http" {
		c.mu.Unlock() // Release lock before calling HTTPClient
		resp, err = c.httpClient.Call(ctx, msg)
		c.mu.Lock() // Reacquire lock
	} else {
		// stdio
		resp, err = c.stdioClient.call(ctx, msg)
	}

	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	c.initialized = true
	return nil
}

// CallTool calls an MCP tool
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not initialized")
	}
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

	var resp *JSONRPCMessage
	var err error

	if c.transportType == "http" {
		result, e := c.httpClient.CallTool(ctx, name, args)
		if e != nil {
			return nil, e
		}
		return result, nil
	}

	resp, err = c.stdioClient.call(ctx, msg)
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
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	if c.transportType == "http" {
		return c.httpClient.ListTools(ctx)
	}

	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not initialized")
	}
	c.mu.RUnlock()

	msg := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextID(),
		Method:  "tools/list",
	}

	resp, err := c.stdioClient.call(ctx, msg)
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
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.ListTools(ctx)
	return err
}

// Close closes the client
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.initialized = false

	if c.httpClient != nil {
		return c.httpClient.Close()
	}

	if c.stdioClient != nil {
		return c.stdioClient.Close()
	}

	return nil
}

// nextID generates the next request ID
func (c *Client) nextID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestID++
	return c.requestID
}
