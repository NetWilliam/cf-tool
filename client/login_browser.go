package client

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/NetWilliam/cf-tool/pkg/logger"
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

	// Navigate to profile page
	color.Cyan("Opening Codeforces profile page in browser...")
	result, err := c.mcpClient.NavigateWithResult(ctx, c.host+"/profile/")
	if err != nil {
		return fmt.Errorf("failed to navigate to profile page: %w", err)
	}

	logger.Debug("Navigate result: %+v", result)

	// Check if already logged in
	color.Cyan("Checking login status...")
	loggedIn, handle, email, err := c.checkProfilePageLoginStatus(ctx)
	if err != nil {
		logger.Warning("Failed to check login status: %v", err)
		// Continue anyway, might need to login
	}

	if loggedIn {
		color.Green("✓ Already logged in as %s\n", handle)
		if email != "" {
			color.Cyan("  Email: %s\n", email)
		}
		c.Handle = handle
		if email != "" {
			c.HandleOrEmail = email
		}
		return c.save()
	}

	// If not logged in, prompt user to login manually
	color.Yellow("\n" +
		"═══════════════════════════════════════════════════════════════\n" +
		"  Please login in the browser within 60 seconds\n" +
		"═══════════════════════════════════════════════════════════════\n\n")

	color.Cyan("Waiting for you to login...")
	if err := c.waitForProfilePageLogin(ctx); err != nil {
		return fmt.Errorf("login timeout: %w", err)
	}

	color.Green("\n✓ Login successful!\n")

	// Get the logged-in handle and email
	_, handle, email, err = c.checkProfilePageLoginStatus(ctx)
	if err != nil {
		// Try to get handle from config as fallback
		if c.HandleOrEmail != "" {
			c.Handle = c.HandleOrEmail
			color.Cyan("Using handle: %s\n", c.Handle)
		}
	} else {
		c.Handle = handle
		if email != "" {
			c.HandleOrEmail = email
			color.Cyan("Email: %s\n", email)
		}
	}

	// Save session
	if err := c.save(); err != nil {
		return err
	}

	return nil
}

// checkProfilePageLoginStatus checks if user is logged in by visiting profile page
func (c *Client) checkProfilePageLoginStatus(ctx context.Context) (loggedIn bool, handle string, email string, err error) {
	// Get page content using HTML content for parsing
	content, err := c.mcpClient.GetWebContentHTML(ctx, c.host+"/profile/")
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get profile page: %w", err)
	}

	logger.Debug("Profile page HTML content length: %d bytes", len(content))

	// Try to find handle in page content
	handle = extractHandleFromProfile(content)
	if handle != "" {
		logger.Debug("Found handle in profile page: %s", handle)
		loggedIn = true
	}

	// Try to extract email from page content
	email = extractEmailFromProfile(content)
	if email != "" {
		logger.Debug("Found email in profile page: %s", email)
	}

	return loggedIn, handle, email, nil
}

// waitForProfilePageLogin waits for user to complete login
func (c *Client) waitForProfilePageLogin(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			loggedIn, handle, email, err := c.checkProfilePageLoginStatus(ctx)
			if err != nil {
				logger.Debug("Error checking login status: %v", err)
				continue // Try again
			}
			if loggedIn {
				c.Handle = handle
				if email != "" {
					c.HandleOrEmail = email
				}
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

// extractHandleFromProfile extracts handle from profile page
// The profile page will redirect to /profile/{HANDLE} if logged in
func extractHandleFromProfile(content string) string {
	// Look for handle in various forms
	// Method 1: Check for handle variable in JavaScript
	reg := regexp.MustCompile(`handle = "([a-zA-Z0-9_\-\.]+)"`)
	matches := reg.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return matches[1]
	}

	// Method 2: Look for profile page URL pattern in content
	// If we see "Profile - Codeforces" or similar patterns
	if strings.Contains(content, "Profile - Codeforces") {
		// Try to extract from meta tags or titles
		titleReg := regexp.MustCompile(`<title>\s*(.+?)\s*-\s*Profile\s*</title>`)
		titleMatches := titleReg.FindStringSubmatch(content)
		if len(titleMatches) >= 2 {
			handle := strings.TrimSpace(titleMatches[1])
			if handle != "" {
				return handle
			}
		}
	}

	// Method 3: Look for profile navigation links
	navReg := regexp.MustCompile(`/profile/([a-zA-Z0-9_\-\.]+)`)
	navMatches := navReg.FindAllStringSubmatch(content, -1)
	if len(navMatches) > 0 {
		// Use the first match as it's likely the current user
		for _, match := range navMatches {
			if len(match) >= 2 {
				candidate := match[1]
				// Filter out common navigation items
				if candidate != "settings" && candidate != "blog" && candidate != "submissions" {
					return candidate
				}
			}
		}
	}

	return ""
}

// extractEmailFromProfile extracts email from profile page
// Looks for: <li><img ... alt="Email" title="Email">USER@gmail.com (invisible)</li>
func extractEmailFromProfile(content string) string {
	// Pattern: Look for <li> containing <img> with alt="Email" and title="Email"
	// followed by email text (with optional parenthetical comment)

	// First, find the <li> elements containing email indicator
	liReg := regexp.MustCompile(`<li[^>]*>.*?<img[^>]*\balt\s*=\s*["']Email["'][^>]*\btitle\s*=\s*["']Email["'][^>]*>(.+?)</li>`)
	matches := liReg.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			emailText := match[1]
			// Extract email, removing parenthetical comments
			email := extractEmailText(emailText)
			if email != "" {
				return email
			}
		}
	}

	// Alternative pattern: Look for email in different formats
	// Some pages might have different HTML structure
	altReg := regexp.MustCompile(`<img[^>]*\balt\s*=\s*["']Email["'][^>]*/?\s*>([^<]+)`)
	altMatches := altReg.FindAllStringSubmatch(content, -1)

	for _, match := range altMatches {
		if len(match) >= 2 {
			emailText := match[1]
			email := extractEmailText(emailText)
			if email != "" {
				return email
			}
		}
	}

	return ""
}

// extractEmailText extracts email from text, removing parenthetical comments and extra whitespace
func extractEmailText(text string) string {
	// Remove HTML entities and tags
	text = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, " ")

	// Remove parenthetical comments like "(invisible)" or "(不可见的)"
	text = regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(text, " ")

	// Clean up whitespace
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Basic email validation
	emailReg := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	email := emailReg.FindString(text)
	return email
}

// checkLoginStatusBrowser checks if the user is logged in via browser (deprecated, use checkProfilePageLoginStatus)
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

// waitForLoginBrowser waits for the user to complete login in the browser (deprecated)
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

// SetMCPClient sets a pre-initialized MCP client and enables browser mode
func (c *Client) SetMCPClient(mcpClient *mcp.Client) {
	c.mcpClient = mcpClient
	c.browserEnabled = true

	// Initialize browser fetcher
	c.fetcher = NewBrowserFetcher(mcpClient)
}

// CloseBrowserClient closes the browser client
func (c *Client) CloseBrowserClient() error {
	if c.mcpClient != nil {
		return c.mcpClient.Close()
	}
	return nil
}
