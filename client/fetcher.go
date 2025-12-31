package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/NetWilliam/cf-tool/pkg/logger"
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
	logger.Info("Initialized HTTPFetcher")
	return &HTTPFetcher{client: client}
}

// Get performs a GET request using util.GetBody
func (f *HTTPFetcher) Get(url string) ([]byte, error) {
	logger.Debug("HTTPFetcher GET: %s", url)

	body, err := util.GetBody(f.client, url)
	if err != nil {
		logger.Error("HTTPFetcher GET failed: %s - %v", url, err)
		return nil, err
	}

	logger.Debug("HTTPFetcher GET success: %s (response size: %d bytes)", url, len(body))
	return body, nil
}

// GetJSON performs a GET request and parses JSON using util.GetJSONBody
func (f *HTTPFetcher) GetJSON(url string) (map[string]interface{}, error) {
	logger.Debug("HTTPFetcher GetJSON: %s", url)

	data, err := util.GetJSONBody(f.client, url)
	if err != nil {
		logger.Error("HTTPFetcher GetJSON failed: %s - %v", url, err)
		return nil, err
	}

	logger.DebugJSON("HTTPFetcher GetJSON response", data)
	return data, nil
}

// Post performs a POST request using util.PostBody
func (f *HTTPFetcher) Post(url string, data url.Values) ([]byte, error) {
	logger.Debug("HTTPFetcher POST: %s", url)
	logger.DebugJSON("HTTPFetcher POST data", data)

	body, err := util.PostBody(f.client, url, data)
	if err != nil {
		logger.Error("HTTPFetcher POST failed: %s - %v", url, err)
		return nil, err
	}

	logger.Debug("HTTPFetcher POST success: %s (response size: %d bytes)", url, len(body))
	return body, nil
}

// BrowserFetcher implements Fetcher using MCP browser client
type BrowserFetcher struct {
	mcpClient *mcp.Client
}

// NewBrowserFetcher creates a new browser fetcher
func NewBrowserFetcher(mcpClient *mcp.Client) *BrowserFetcher {
	logger.Info("Initialized BrowserFetcher")
	return &BrowserFetcher{mcpClient: mcpClient}
}

// Get performs a GET request using browser
func (f *BrowserFetcher) Get(url string) ([]byte, error) {
	logger.Debug("BrowserFetcher GET: %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout*1000000000) // convert to nanoseconds
	defer cancel()

	// Use GetWebContentHTML to get raw HTML instead of text content
	content, err := f.mcpClient.GetWebContentHTML(ctx, url)
	if err != nil {
		logger.Error("BrowserFetcher GET failed: %s - %v", url, err)
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	logger.Debug("BrowserFetcher GET success: %s (response size: %d bytes)", url, len(content))
	return []byte(content), nil
}

// GetJSON performs a GET request using browser and parses JSON
func (f *BrowserFetcher) GetJSON(url string) (map[string]interface{}, error) {
	logger.Debug("BrowserFetcher GetJSON: %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout*1000000000)
	defer cancel()

	content, err := f.mcpClient.GetWebContent(ctx, url)
	if err != nil {
		logger.Error("BrowserFetcher GetJSON failed: %s - %v", url, err)
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	// Parse JSON from content
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		logger.Error("BrowserFetcher GetJSON parse failed: %s - %v", url, err)
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	logger.DebugJSON("BrowserFetcher GetJSON response", result)
	return result, nil
}

// Post performs a POST request using browser
func (f *BrowserFetcher) Post(url string, data url.Values) ([]byte, error) {
	logger.Debug("BrowserFetcher POST: %s", url)
	logger.DebugJSON("BrowserFetcher POST data", data)

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
		logger.Error("BrowserFetcher POST failed: %s - %v", url, err)
		return nil, fmt.Errorf("browser request failed: %w", err)
	}

	// Extract response data
	if len(result.Content) > 0 {
		if text, ok := result.Content[0].(map[string]interface{}); ok {
			if content, ok := text["text"].(string); ok {
				logger.Debug("BrowserFetcher POST success: %s (response size: %d bytes)", url, len(content))
				return []byte(content), nil
			}
			if data, ok := text["data"].(string); ok {
				logger.Debug("BrowserFetcher POST success: %s (response size: %d bytes)", url, len(data))
				return []byte(data), nil
			}
		}
	}

	logger.Error("BrowserFetcher POST failed: %s - failed to extract response", url)
	return nil, fmt.Errorf("failed to extract response from browser")
}

const defaultTimeout = 30 // seconds

