// Package mcp implements a Model Context Protocol (MCP) client for communicating
// with browser automation servers like MCP-Chrome.
//
// This package provides:
//   - JSON-RPC 2.0 communication over stdio
//   - Tool invocation interface
//   - Chrome-specific helper methods
//
// Basic usage:
//
//	client, err := mcp.NewClient("node", []string{"/path/to/server.js"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Navigate to a URL
//	err = client.Navigate(ctx, "https://example.com")
//
//	// Get page content
//	content, err := client.GetWebContent(ctx, "https://example.com")
package mcp
