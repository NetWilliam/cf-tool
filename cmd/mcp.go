package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/NetWilliam/cf-tool/pkg/mcp"
	"github.com/fatih/color"
)

// McpPing tests the MCP Chrome server connection
func McpPing() error {
	color.Cyan("Testing MCP Chrome server connection...\n")

	// Try to find MCP server (HTTP or stdio)
	serverURL, mcpPath, err := findMCPServer()
	if err != nil {
		color.Red("❌ MCP server not found: %v", err)
		printInstallationHints()
		return err
	}

	var mcpClient *mcp.Client

	// Determine which transport to use
	if serverURL != "" {
		color.White("Using HTTP transport: %s\n", serverURL)
		mcpClient, err = mcp.NewClientHTTP(serverURL)
	} else if mcpPath != "" {
		color.White("Using stdio transport: %s\n", mcpPath)
		// For stdio, we need node and the script path
		mcpClient, err = mcp.NewClient("node", []string{mcpPath})
	} else {
		color.Red("❌ No valid MCP server configuration found")
		printInstallationHints()
		return fmt.Errorf("no MCP server found")
	}

	if err != nil {
		color.Red("❌ Failed to create MCP client: %v", err)
		printInstallationHints()
		return err
	}
	defer mcpClient.Close()

	// Ping test
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	color.Cyan("Pinging MCP server...")
	if err := mcpClient.Ping(ctx); err != nil {
		color.Red("❌ MCP server ping failed: %v", err)
		printInstallationHints()
		return err
	}

	color.Green("✓ MCP server is running!\n")

	// Get available tools
	color.Cyan("Listing available tools...")
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		color.Yellow("⚠ Could not list tools: %v", err)
		return nil
	}

	color.Cyan("\nAvailable tools (%d):\n", len(tools))
	for i, tool := range tools {
		color.White("  %d. %s", i+1, tool.Name)
		if tool.Description != "" {
			color.White("     %s", tool.Description)
		}
	}

	// Check for Chrome-specific tools
	chromeTools := []string{
		"chrome_navigate",
		"chrome_get_web_content",
		"chrome_network_request",
		"chrome_fill_or_select",
		"chrome_click_element",
	}

	color.Cyan("\nChecking Chrome integration...")
	missing := []string{}
	for _, toolName := range chromeTools {
		found := false
		for _, tool := range tools {
			if tool.Name == toolName {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, toolName)
		}
	}

	if len(missing) > 0 {
		color.Yellow("⚠ Warning: Some Chrome tools are missing:")
		for _, name := range missing {
			color.Yellow("   - %s", name)
		}
	} else {
		color.Green("✓ All Chrome tools are available!")
	}

	color.Green("\n✓ Your browser is ready to use with CF-Tool!")
	return nil
}

// findMCPServer tries to locate the MCP server (HTTP or stdio)
func findMCPServer() (serverURL string, mcpPath string, err error) {
	// Check environment variable first
	if envURL := os.Getenv("MCP_SERVER_URL"); envURL != "" {
		if _, parseErr := url.Parse(envURL); parseErr == nil {
			return envURL, "", nil
		}
	}

	// Try default HTTP port with /ping endpoint
	pingURL := "http://127.0.0.1:12306/ping"

	// Quick test if server is available
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try to ping the server
	req, reqErr := http.NewRequestWithContext(ctx, "GET", pingURL, nil)
	if reqErr == nil {
		resp, httpErr := http.DefaultClient.Do(req)
		if httpErr == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				// Server is alive, return the MCP endpoint
				return "http://127.0.0.1:12306/mcp", "", nil
			}
		}
	}

	// Fall back to stdio paths
	possiblePaths := []string{
		os.Getenv("HOME") + "/.mcp-chrome/mcp-chrome/app/native-server/dist/mcp/mcp-server-stdio.js",
		os.Getenv("HOME") + "/.mcp-chrome/mcp-chrome-bridge/dist/mcp/mcp-server-stdio.js",
		os.Getenv("HOME") + "/.local/share/mcp-chrome/dist/mcp/mcp-server-stdio.js",
	}

	// Check if any path exists
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return "", path, nil
		}
	}

	return "", "", fmt.Errorf("MCP server not found")
}

// printInstallationHints prints installation instructions
func printInstallationHints() {
	color.Cyan("\n📦 Installation Guide:\n")
	color.White(`
1. Install Chrome Extension:
   - Download from: https://github.com/hangwin/mcp-chrome/releases
   - Or clone: git clone https://github.com/hangwin/mcp-chrome.git ~/.mcp-chrome/mcp-chrome
   - Open Chrome: chrome://extensions/
   - Enable "Developer mode"
   - Click "Load unpacked"
   - Select: ~/.mcp-chrome/mcp-chrome/app/chrome-extension

2. Install Native Host:
   $ cd ~/.mcp-chrome/mcp-chrome/app/native-server
   $ pnpm install
   $ pnpm build
   $ pnpm run register

3. Configure CF Tool:
   Edit ~/.cf/config and set:
   {
     "browser": {
       "enabled": true,
       "mcp_transport": "http",
       "mcp_server_url": "http://127.0.0.1:12306/mcp"
     }
   }

4. Verify Installation:
   $ cf mcp-ping

For more details, visit: https://github.com/hangwin/mcp-chrome
`)
}
