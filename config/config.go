package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// CodeTemplate config parse code template
type CodeTemplate struct {
	Alias        string   `json:"alias"`
	Lang         string   `json:"lang"`
	Path         string   `json:"path"`
	Suffix       []string `json:"suffix"`
	BeforeScript string   `json:"before_script"`
	Script       string   `json:"script"`
	AfterScript  string   `json:"after_script"`
}

// Config load and save configuration
type Config struct {
	Template      []CodeTemplate    `json:"template"`
	Default       int               `json:"default"`
	GenAfterParse bool              `json:"gen_after_parse"`
	Host          string            `json:"host"`
	Proxy         string            `json:"proxy"`
	FolderName    map[string]string `json:"folder_name"`
	Browser       BrowserConfig     `json:"browser"`
	path          string
}

// BrowserConfig configuration for browser mode
type BrowserConfig struct {
	// Whether browser mode is enabled
	Enabled bool `json:"enabled"`

	// MCP transport type: "stdio" or "http"
	Transport string `json:"mcp_transport"`

	// MCP server command (for stdio transport)
	Command string `json:"mcp_command"`

	// MCP server arguments (for stdio transport)
	Args []string `json:"mcp_args"`

	// MCP server URL (for HTTP transport)
	ServerURL string `json:"mcp_server_url"`

	// Auto login (false = manual login in browser)
	AutoLogin bool `json:"auto_login"`

	// Fallback to HTTP if browser fails
	FallbackToHTTP bool `json:"fallback_to_http"`
}

// Instance global configuration
var Instance *Config

// Init initialize
func Init(path string) {
	c := &Config{
		path:   path,
		Host:   "https://codeforces.com",
		Proxy:  "",
		Browser: BrowserConfig{
			Enabled:        false,
			Transport:      "http",
			Command:        "node",
			Args:           []string{},
			ServerURL:      "http://127.0.0.1:12306/mcp",
			AutoLogin:      false,
			FallbackToHTTP: false,
		},
	}
	if err := c.load(); err != nil {
		color.Red(err.Error())
		color.Green("Create a new configuration in %v", path)
	}
	if c.Default < 0 || c.Default >= len(c.Template) {
		c.Default = 0
	}
	if c.FolderName == nil {
		c.FolderName = map[string]string{}
	}
	if _, ok := c.FolderName["root"]; !ok {
		c.FolderName["root"] = "cf"
	}
	// Default problem types
	problemTypes := []string{"contest", "gym", "problemset", "edu"}
	for _, problemType := range problemTypes {
		if _, ok := c.FolderName[problemType]; !ok {
			c.FolderName[problemType] = problemType
		}
	}
	c.save()
	Instance = c
}

// load from path
func (c *Config) load() (err error) {
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
func (c *Config) save() (err error) {
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(c)
	if err == nil {
		os.MkdirAll(filepath.Dir(c.path), os.ModePerm)
		err = os.WriteFile(c.path, data.Bytes(), 0644)
	}
	if err != nil {
		color.Red("Cannot save config to %v\n%v", c.path, err.Error())
	}
	return
}
