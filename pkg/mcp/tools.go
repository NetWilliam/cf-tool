package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolNames defines constants for MCP Chrome tools
type ToolNames struct {
	GetWindowsAndTabs     string
	Navigate              string
	GetWebContent         string
	NetworkRequest        string
	NetworkCaptureStart   string
	NetworkCaptureStop    string
	FillOrSelect          string
	ClickElement          string
	Keyboard              string
}

// ChromeTools contains Chrome MCP tool names
var ChromeTools = ToolNames{
	GetWindowsAndTabs:  "get_windows_and_tabs",
	Navigate:           "chrome_navigate",
	GetWebContent:      "chrome_get_web_content",
	NetworkRequest:     "chrome_network_request",
	NetworkCaptureStart: "chrome_network_capture_start",
	NetworkCaptureStop: "chrome_network_capture_stop",
	FillOrSelect:       "chrome_fill_or_select",
	ClickElement:       "chrome_click_element",
	Keyboard:           "chrome_keyboard",
}

// Helper functions for common tool calls

// Navigate navigates to a URL
func (c *Client) Navigate(ctx context.Context, url string) error {
	_, err := c.CallTool(ctx, ChromeTools.Navigate, map[string]interface{}{
		"url": url,
	})
	return err
}

// NavigateWithResult navigates to a URL and returns the result
func (c *Client) NavigateWithResult(ctx context.Context, url string) (*ToolResult, error) {
	return c.CallTool(ctx, ChromeTools.Navigate, map[string]interface{}{
		"url": url,
	})
}

// GetWebContent gets the text content of a page
func (c *Client) GetWebContent(ctx context.Context, url string) (string, error) {
	result, err := c.CallTool(ctx, ChromeTools.GetWebContent, map[string]interface{}{
		"url":         url,
		"textContent": true,
	})
	if err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content returned")
	}

	// Extract text from content
	if text, ok := result.Content[0].(map[string]interface{}); ok {
		if content, ok := text["text"].(string); ok {
			return content, nil
		}
	}

	return "", fmt.Errorf("unexpected content format")
}

// GetWebContentHTML gets the HTML content of a page
func (c *Client) GetWebContentHTML(ctx context.Context, url string) (string, error) {
	result, err := c.CallTool(ctx, ChromeTools.GetWebContent, map[string]interface{}{
		"url":         url,
		"htmlContent": true,
	})
	if err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content returned")
	}

	// Extract HTML from content - the chrome_get_web_content tool returns JSON
	if html, ok := result.Content[0].(map[string]interface{}); ok {
		// Try "html" field first
		if content, ok := html["html"].(string); ok {
			return content, nil
		}
		// Try "htmlContent" field (chrome_get_web_content uses this name)
		if content, ok := html["htmlContent"].(string); ok {
			return content, nil
		}
		// Try "text" field - it might contain JSON that needs parsing
		if textContent, ok := html["text"].(string); ok {
			// Check if it's JSON-formatted
			if len(textContent) > 0 && textContent[0] == '{' {
				// Parse as JSON to extract htmlContent
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(textContent), &jsonData); err == nil {
					if htmlContent, ok := jsonData["htmlContent"].(string); ok {
						return htmlContent, nil
					}
				}
			}
			// Otherwise return as-is
			return textContent, nil
		}
		// Log what we got for debugging
		return "", fmt.Errorf("unexpected content format, got keys: %v", html)
	}

	return "", fmt.Errorf("unexpected content format: %T", result.Content[0])
}

// NetworkRequest sends a network request
type NetworkRequestOptions struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    interface{}
}

func (c *Client) NetworkRequest(ctx context.Context, opts *NetworkRequestOptions) (*ToolResult, error) {
	args := map[string]interface{}{
		"url": opts.URL,
	}

	if opts.Method != "" {
		args["method"] = opts.Method
	}
	if opts.Headers != nil {
		args["headers"] = opts.Headers
	}
	if opts.Body != nil {
		args["body"] = opts.Body
	}

	return c.CallTool(ctx, ChromeTools.NetworkRequest, args)
}

// Fill fills a form element
func (c *Client) Fill(ctx context.Context, selector, value string) error {
	_, err := c.CallTool(ctx, ChromeTools.FillOrSelect, map[string]interface{}{
		"selector": selector,
		"value":    value,
	})
	return err
}

// Click clicks an element
func (c *Client) Click(ctx context.Context, selector string) error {
	_, err := c.CallTool(ctx, ChromeTools.ClickElement, map[string]interface{}{
		"selector": selector,
	})
	return err
}

// Keyboard sends keyboard input
func (c *Client) Keyboard(ctx context.Context, keys string) error {
	_, err := c.CallTool(ctx, ChromeTools.Keyboard, map[string]interface{}{
		"keys": keys,
	})
	return err
}
