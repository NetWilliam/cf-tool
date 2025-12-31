package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/NetWilliam/cf-tool/pkg/logger"
	"github.com/NetWilliam/cf-tool/cookiejar"
	"github.com/NetWilliam/cf-tool/pkg/mcp"
)

// Client codeforces client
type Client struct {
	Jar            *cookiejar.Jar `json:"cookies"`
	Handle         string         `json:"handle"`
	HandleOrEmail  string         `json:"handle_or_email"`
	Password       string         `json:"password"`
	Ftaa           string         `json:"ftaa"`
	Bfaa           string         `json:"bfaa"`
	LastSubmission *Info          `json:"last_submission"`
	host           string
	proxy          string
	path           string
	client         *http.Client
	mcpClient      *mcp.Client    `json:"-"`      // MCP client for browser mode
	browserEnabled bool           `json:"-"`      // Whether browser mode is enabled
	fetcher        Fetcher        `json:"-"`      // Unified fetcher interface
}

// Instance global client
var Instance *Client

// Init initialize
func Init(path, host, proxy string) {
	// Check for CF_DEBUG environment variable
	if os.Getenv("CF_DEBUG") != "" {
		logger.SetLevel(logger.DebugLevel)
	}

	jar, _ := cookiejar.New(nil)
	c := &Client{Jar: jar, LastSubmission: nil, path: path, host: host, proxy: proxy, client: nil, browserEnabled: false}
	if err := c.load(); err != nil {
		color.Red(err.Error())
		color.Green("Create a new session in %v", path)
	}
	Proxy := http.ProxyFromEnvironment
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			color.Red(err.Error())
			color.Green("Use default proxy from environment")
		} else {
			Proxy = http.ProxyURL(proxyURL)
		}
	}
	c.client = &http.Client{Jar: c.Jar, Transport: &http.Transport{Proxy: Proxy}}

	// Try to initialize browser mode (will auto-detect MCP server)
	logger.Info("Initializing browser mode...\n")
	if err := c.initBrowserMode(); err != nil {
		logger.Warning("Browser mode not available: %v\n", err)
		logger.Info("Falling back to HTTP mode. Some features may not work.\n")
		c.fetcher = NewHTTPFetcher(c.client)
	}

	// Try to load user info from profile page
    /*
	if c.browserEnabled {
		c.loadUserInfoFromBrowser()
	}
    */

	if err := c.save(); err != nil {
		color.Red(err.Error())
	}
	Instance = c
}

// initBrowserMode attempts to initialize browser mode by detecting MCP server
func (c *Client) initBrowserMode() error {
	// Try to find MCP server
	serverURL, mcpPath, err := findMCPServer()
	if err != nil {
		return fmt.Errorf("MCP server not found: %w", err)
	}

	var mcpClient *mcp.Client

	// Determine which transport to use
	if serverURL != "" {
		logger.Info("Using HTTP transport: %s\n", serverURL)
		mcpClient, err = mcp.NewClientHTTP(serverURL)
	} else if mcpPath != "" {
		logger.Info("Using stdio transport: %s\n", mcpPath)
		mcpClient, err = mcp.NewClient("node", []string{mcpPath})
	}

	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}

	// Set MCP client and enable browser mode
	c.SetMCPClient(mcpClient)
	logger.Info("✓ Browser mode enabled\n")

	return nil
}

// findMCPServer attempts to find MCP Chrome Server
func findMCPServer() (serverURL, mcpPath string, err error) {
	// Try HTTP transport first (check environment variable)
	if envURL := os.Getenv("CF_MCP_HTTP_URL"); envURL != "" {
		return envURL, "", nil
	}

	// Default HTTP MCP URL
	if _, err := http.Get("http://127.0.0.1:12306/mcp"); err == nil {
		return "http://127.0.0.1:12306/mcp", "", nil
	}

	// TODO: Add stdio transport detection if needed

	return "", "", fmt.Errorf("no MCP server found")
}

// loadUserInfoFromBrowser loads user handle and email from profile page
func (c *Client) loadUserInfoFromBrowser() {
	logger.Info("Loading user info from browser...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	content, err := c.mcpClient.GetWebContentHTML(ctx, c.host+"/profile/")
	if err != nil {
		color.Yellow("Failed to load profile page: %v\n", err)
		color.Cyan("You can manually set your handle in config if needed\n")
		return
	}

	// Extract handle
	handle := extractHandleFromProfile(content)
	if handle != "" {
		c.Handle = handle
		color.Green("✓ Handle: %s\n", handle)
	}

	// Extract email
	email := extractEmailFromProfile(content)
	if email != "" {
		c.HandleOrEmail = email
		logger.Info("Email: %s\n", email)
	}

	if handle == "" && email == "" {
		color.Yellow("Could not extract user info from profile page\n")
		color.Cyan("Please make sure you're logged in to Codeforces in your browser\n")
	}
}

// load from path
func (c *Client) load() (err error) {
	file, err := os.Open(c.path)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)

	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, c)
}

// save file to path
func (c *Client) save() (err error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		os.MkdirAll(filepath.Dir(c.path), os.ModePerm)
		err = os.WriteFile(c.path, data, 0644)
	}
	if err != nil {
		color.Red("Cannot save session to %v\n%v", c.path, err.Error())
	}
	return
}
