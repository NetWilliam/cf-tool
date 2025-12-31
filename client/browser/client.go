package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/NetWilliam/cf-tool/pkg/mcp"
)

// Client represents a browser-based HTTP client
type Client struct {
	mcpClient    *mcp.Client
	timeout      time.Duration
	autoRedirect bool
}

// NewClient creates a new browser HTTP client
func NewClient(mcpClient *mcp.Client) *Client {
	return &Client{
		mcpClient:    mcpClient,
		timeout:      30 * time.Second,
		autoRedirect: true,
	}
}

// SetTimeout sets the request timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Get performs a GET request and returns the response body
func (c *Client) Get(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	result, err := c.mcpClient.NetworkRequest(ctx, &mcp.NetworkRequestOptions{
		URL:    url,
		Method: "GET",
	})

	if err != nil {
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	return c.extractResponseData(result)
}

// Post performs a POST request with form data
func (c *Client) Post(url string, data url.Values) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	result, err := c.mcpClient.NetworkRequest(ctx, &mcp.NetworkRequestOptions{
		URL:    url,
		Method: "POST",
		Body:   data.Encode(),
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	})

	if err != nil {
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	return c.extractResponseData(result)
}

// PostJSON performs a POST request with JSON body
func (c *Client) PostJSON(url string, data interface{}) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	result, err := c.mcpClient.NetworkRequest(ctx, &mcp.NetworkRequestOptions{
		URL:    url,
		Method: "POST",
		Body:   string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})

	if err != nil {
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	return c.extractResponseData(result)
}

// GetJSON performs a GET request and parses JSON response
func (c *Client) GetJSON(url string) (map[string]interface{}, error) {
	body, err := c.Get(url)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// GetContent gets the text content of a web page
func (c *Client) GetContent(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	content, err := c.mcpClient.GetWebContent(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to get web content: %w", err)
	}

	return content, nil
}

// Do performs an arbitrary HTTP request
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var body interface{}
	if req.Body != nil {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		body = string(data)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Build headers
	headers := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	result, err := c.mcpClient.NetworkRequest(ctx, &mcp.NetworkRequestOptions{
		URL:     req.URL.String(),
		Method:  req.Method,
		Headers: headers,
		Body:    body,
	})

	if err != nil {
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	// Convert to http.Response
	return c.buildResponse(req, result)
}

// extractResponseData extracts raw data from tool result
func (c *Client) extractResponseData(result *mcp.ToolResult) ([]byte, error) {
	if result.IsError {
		// Try to extract error message
		if len(result.Content) > 0 {
			if text, ok := result.Content[0].(map[string]interface{}); ok {
				if errMsg, ok := text["text"].(string); ok {
					return nil, fmt.Errorf("request error: %s", errMsg)
				}
			}
		}
		return nil, fmt.Errorf("request failed")
	}

	// Extract data from content
	if len(result.Content) == 0 {
		return []byte{}, nil
	}

	// Handle text content
	if text, ok := result.Content[0].(map[string]interface{}); ok {
		if content, ok := text["text"].(string); ok {
			return []byte(content), nil
		}
		if content, ok := text["data"].(string); ok {
			return []byte(content), nil
		}
	}

	// Try to marshal to JSON and return
	data, err := json.Marshal(result.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to encode response: %w", err)
	}

	return data, nil
}

// buildResponse builds an http.Response from tool result
func (c *Client) buildResponse(req *http.Request, result *mcp.ToolResult) (*http.Response, error) {
	data, err := c.extractResponseData(result)
	if err != nil {
		return nil, err
	}

	// Build a minimal http.Response
	resp := &http.Response{
		Request:    req,
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(string(data))),
		Header:     make(http.Header),
	}

	// Set content type
	resp.Header.Set("Content-Type", "text/html; charset=utf-8")

	return resp, nil
}
