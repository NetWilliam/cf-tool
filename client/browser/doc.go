// Package browser implements a browser-based HTTP client using MCP Chrome.
//
// This package provides HTTP client functionality that routes requests through
// the user's browser, bypassing Cloudflare protections and other anti-bot measures.
//
// Basic usage:
//
//	mcpClient, err := mcp.NewClient("node", []string{"/path/to/server.js"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mcpClient.Close()
//
//	browserClient := browser.NewClient(mcpClient)
//
//	// Perform GET request
//	body, err := browserClient.Get("https://codeforces.com")
//
//	// Perform POST request
//	data := url.Values{}
//	data.Set("key", "value")
//	body, err = browserClient.Post("https://example.com", data)
//
// The browser client automatically handles:
//   - Cookies from the browser session
//   - JavaScript challenges (Cloudflare)
//   - User agent and headers
//   - Redirects
package browser
