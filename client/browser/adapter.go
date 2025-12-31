package browser

import (
	"net/http"

	"github.com/NetWilliam/cf-tool/pkg/mcp"
)

// HTTPClient interface mimics net/http.Client for compatibility
type HTTPClient interface {
	Get(url string) (*http.Response, error)
	Post(url string, contentType string, body interface{}) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

// BrowserAdapter adapts BrowserClient to http.Client interface
type BrowserAdapter struct {
	browser *Client
}

// NewBrowserAdapter creates a new browser adapter
func NewBrowserAdapter(mcpClient *mcp.Client) *BrowserAdapter {
	return &BrowserAdapter{
		browser: NewClient(mcpClient),
	}
}

// Get performs a GET request
func (a *BrowserAdapter) Get(url string) (*http.Response, error) {
	body, err := a.browser.Get(url)
	if err != nil {
		return nil, err
	}

	return buildResponse(url, body), nil
}

// Post performs a POST request
func (a *BrowserAdapter) Post(url string, contentType string, body interface{}) (*http.Response, error) {
	// For now, we only handle form data
	return nil, nil
}

// Do performs an HTTP request
func (a *BrowserAdapter) Do(req *http.Request) (*http.Response, error) {
	return a.browser.Do(req)
}

// buildResponse builds an http.Response from raw body
func buildResponse(url string, body []byte) *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       &stringReadCloser{strings.NewReader(string(body))},
		Header:     make(http.Header),
	}
}

// stringReadCloser wraps io.StringReader to implement io.ReadCloser
type stringReadCloser struct {
	*strings.Reader
}

// Close implements io.ReadCloser
func (rc *stringReadCloser) Close() error {
	return nil
}
