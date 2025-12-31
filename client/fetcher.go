package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/NetWilliam/cf-tool/pkg/mcp"
	"github.com/NetWilliam/cf-tool/util"
)

// Fetcher defines the interface for fetching data from different sources
type Fetcher interface {
	// Get performs a GET request and returns the response body
	Get(url string) ([]byte, error)

	// GetJSON performs a GET request and parses JSON response
	GetJSON(url string) (map[string]interface{}, error)

	// Post performs a POST request with form data
	Post(url string, data url.Values) ([]byte, error)
}

// HTTPFetcher implements Fetcher using standard HTTP client
type HTTPFetcher struct {
	client *http.Client
}

// NewHTTPFetcher creates a new HTTP fetcher
func NewHTTPFetcher(client *http.Client) *HTTPFetcher {
	return &HTTPFetcher{client: client}
}

// Get performs a GET request using util.GetBody
func (f *HTTPFetcher) Get(url string) ([]byte, error) {
	return util.GetBody(f.client, url)
}

// GetJSON performs a GET request and parses JSON using util.GetJSONBody
func (f *HTTPFetcher) GetJSON(url string) (map[string]interface{}, error) {
	return util.GetJSONBody(f.client, url)
}

// Post performs a POST request using util.PostBody
func (f *HTTPFetcher) Post(url string, data url.Values) ([]byte, error) {
	return util.PostBody(f.client, url, data)
}

// BrowserFetcher implements Fetcher using MCP browser client
type BrowserFetcher struct {
	mcpClient *mcp.Client
}

// NewBrowserFetcher creates a new browser fetcher
func NewBrowserFetcher(mcpClient *mcp.Client) *BrowserFetcher {
	return &BrowserFetcher{mcpClient: mcpClient}
}

// Get performs a GET request using browser
func (f *BrowserFetcher) Get(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout*1000000000) // convert to nanoseconds
	defer cancel()

	content, err := f.mcpClient.GetWebContent(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	return []byte(content), nil
}

// GetJSON performs a GET request using browser and parses JSON
func (f *BrowserFetcher) GetJSON(url string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout*1000000000)
	defer cancel()

	content, err := f.mcpClient.GetWebContent(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	// Parse JSON from content
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// Post performs a POST request using browser
func (f *BrowserFetcher) Post(url string, data url.Values) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout*1000000000)
	defer cancel()

	// Use chrome_network_request for POST
	result, err := f.mcpClient.NetworkRequest(ctx, &mcp.NetworkRequestOptions{
		URL:     url,
		Method:  "POST",
		Body:    data.Encode(),
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	})

	if err != nil {
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	// Extract response data
	if len(result.Content) > 0 {
		if text, ok := result.Content[0].(map[string]interface{}); ok {
			if content, ok := text["text"].(string); ok {
				return []byte(content), nil
			}
			if data, ok := text["data"].(string); ok {
				return []byte(data), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to extract response from browser")
}

const defaultTimeout = 30 // seconds

