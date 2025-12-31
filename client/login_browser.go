package client

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/NetWilliam/cf-tool/pkg/mcp"
	"github.com/fatih/color"
)

// LoginWithBrowser performs login using the browser
func (c *Client) LoginWithBrowser() error {
	color.Cyan("Browser mode login...\n")

	// Check if MCP client is initialized
	if c.mcpClient == nil {
		return fmt.Errorf("MCP client not initialized. Please enable browser mode in config and restart")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Navigate to login page
	color.Cyan("Opening Codeforces login page in browser...")
	if err := c.mcpClient.Navigate(ctx, c.host+"/enter"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Check if already logged in
	color.Cyan("Checking login status...")
	loggedIn, handle, err := c.checkLoginStatusBrowser(ctx)
	if err != nil {
		return fmt.Errorf("failed to check login status: %w", err)
	}

	if loggedIn {
		color.Green("✓ Already logged in as %s\n", handle)
		c.Handle = handle
		return c.save()
	}

	// If not logged in, prompt user to login manually
	color.Yellow("\n" +
		"═══════════════════════════════════════════════════════════════\n" +
		"  Please login in the browser within 60 seconds\n" +
		"═══════════════════════════════════════════════════════════════\n\n")

	color.Cyan("Waiting for you to login...")
	if err := c.waitForLoginBrowser(ctx); err != nil {
		return fmt.Errorf("login timeout: %w", err)
	}

	color.Green("\n✓ Login successful!\n")

	// Get the logged-in handle
	_, handle, err = c.checkLoginStatusBrowser(ctx)
	if err != nil {
		// Try to get handle from config as fallback
		if c.HandleOrEmail != "" {
			c.Handle = c.HandleOrEmail
			color.Cyan("Using handle: %s\n", c.Handle)
		}
	} else {
		c.Handle = handle
	}

	// Save session
	if err := c.save(); err != nil {
		return err
	}

	return nil
}

// checkLoginStatusBrowser checks if the user is logged in via browser
func (c *Client) checkLoginStatusBrowser(ctx context.Context) (bool, string, error) {
	content, err := c.mcpClient.GetWebContent(ctx, c.host+"/enter")
	if err != nil {
		return false, "", fmt.Errorf("failed to get page content: %w", err)
	}

	handle, err := findHandleFromContent([]byte(content))
	if err != nil {
		return false, "", nil // Not logged in, but not an error
	}

	return true, handle, nil
}

// waitForLoginBrowser waits for the user to complete login in the browser
func (c *Client) waitForLoginBrowser(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			loggedIn, handle, err := c.checkLoginStatusBrowser(ctx)
			if err != nil {
				continue // Try again
			}
			if loggedIn {
				c.Handle = handle
				return nil
			}
			color.White(".")
		case <-timeout:
			return fmt.Errorf("login timeout (60 seconds)")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// findHandleFromContent extracts handle from HTML content
func findHandleFromContent(body []byte) (string, error) {
	reg := regexp.MustCompile(`handle = "([\s\S]+?)"`)
	tmp := reg.FindSubmatch(body)
	if len(tmp) < 2 {
		return "", errors.New(ErrorNotLogged)
	}
	return string(tmp[1]), nil
}

// InitBrowserClient initializes the MCP browser client with given parameters
func (c *Client) InitBrowserClient(command string, args []string) error {
	mcpClient, err := mcp.NewClient(command, args)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}

	c.mcpClient = mcpClient
	c.browserEnabled = true

	// Initialize browser fetcher
	c.fetcher = NewBrowserFetcher(mcpClient)

	return nil
}

// CloseBrowserClient closes the browser client
func (c *Client) CloseBrowserClient() error {
	if c.mcpClient != nil {
		return c.mcpClient.Close()
	}
	return nil
}
