package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SimpleHTTPClient is a minimal HTTP client for MCP operations
type SimpleHTTPClient struct {
	serverURL string
	client    *http.Client
}

// NewSimpleHTTPClient creates a new simple HTTP client
func NewSimpleHTTPClient(serverURL string) *SimpleHTTPClient {
	return &SimpleHTTPClient{
		serverURL: serverURL,
		client:    &http.Client{},
	}
}

// Post sends a JSON-RPC request and returns the response
func (c *SimpleHTTPClient) Post(ctx context.Context, msg *JSONRPCMessage) (*JSONRPCMessage, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(data))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var response JSONRPCMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// StdioClient handles stdio-based MCP communication
type StdioClient struct {
	transport *StdioTransport
}

// NewStdioClient creates a new stdio client
func NewStdioClient(command string, args []string) (*StdioClient, error) {
	transport, err := NewStdioTransport(command, args)
	if err != nil {
		return nil, err
	}

	return &StdioClient{
		transport: transport,
	}, nil
}

// Call sends a JSON-RPC message and waits for response
func (c *StdioClient) call(ctx context.Context, msg *JSONRPCMessage) (*JSONRPCMessage, error) {
	// Send request
	if err := c.transport.Send(ctx, msg); err != nil {
		return nil, err
	}

	// Wait for response with matching ID
	for i := 0; i < 50; i++ {
		resp, err := c.transport.Receive(ctx)
		if err != nil {
			return nil, err
		}

		// Check if this is a response to our request
		if resp.ID != nil && resp.ID == msg.ID {
			return resp, nil
		}

		// Handle notifications (messages without ID)
		if resp.ID == nil {
			continue
		}
	}

	return nil, fmt.Errorf("timeout waiting for response")
}

// Close closes the stdio client
func (c *StdioClient) Close() error {
	return c.transport.Close()
}
