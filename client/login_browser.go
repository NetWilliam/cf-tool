package client

import (
	"regexp"
	"strings"

	"github.com/NetWilliam/cf-tool/pkg/mcp"
)

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
