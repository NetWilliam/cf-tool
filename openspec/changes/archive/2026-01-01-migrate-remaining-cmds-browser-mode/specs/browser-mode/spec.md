## ADDED Requirements

### Requirement: Browser Mode Architecture

The system SHALL use browser automation via mcp-chrome (Chrome DevTools Protocol) for all network-dependent operations. The system SHALL NOT provide HTTP mode fallback for commands that interact with Codeforces.

#### Scenario: Initialize browser mode on startup

- **GIVEN** cf-tool is starting up
- **WHEN** the client is initialized
- **THEN** the system attempts to connect to mcp-chrome-bridge
- **AND** if connection succeeds, browser mode is enabled
- **AND** if connection fails, the system returns an error
- **AND** the error message instructs user to install mcp-chrome
- **AND** the system does NOT fall back to HTTP mode

#### Scenario: Network commands use BrowserFetcher

- **GIVEN** browser mode is enabled successfully
- **WHEN** any network-dependent command is executed (list, parse, submit, watch, race, pull, clone)
- **THEN** the command uses BrowserFetcher for all HTTP requests
- **AND** BrowserFetcher uses mcp-chrome to execute requests in Chrome browser
- **AND** Cloudflare protection is bypassed automatically
- **AND** no HTTP client is used for Codeforces requests

#### Scenario: Local commands bypass browser mode

- **GIVEN** browser mode may or may not be enabled
- **WHEN** local-only commands are executed (config, test, gen, open, stand, sid)
- **THEN** the commands execute without requiring browser mode
- **AND** file operations (config, test, gen) work normally
- **AND** URL opening (open, stand, sid) uses system default browser
- **AND** no mcp-chrome connection is attempted

### Requirement: Browser Mode Error Handling

The system SHALL provide clear, actionable error messages when browser mode is unavailable. Error messages SHALL guide users through installation and troubleshooting steps.

#### Scenario: Clear error message on browser unavailable

- **GIVEN** browser mode is not available (mcp-chrome not installed or bridge not running)
- **WHEN** user runs a network-dependent command
- **THEN** the system displays error message with:
  - Emoji indicator: ❌
  - Clear statement: "Browser mode is required but not available"
  - Numbered installation steps (1-4)
  - Link to mcp-chrome GitHub releases
  - Command to install bridge: `npm install -g @hangwin/mcp-chrome-bridge`
  - Command to start bridge: `mcp-chrome-bridge`
  - Command to verify: `cf mcp-ping`
  - Reference to README.md
- **AND** the error message is formatted for readability (line breaks, indentation)

#### Scenario: Helpful error for connection failure

- **GIVEN** mcp-chrome-bridge is not running
- **WHEN** cf-tool attempts to initialize browser mode
- **THEN** the error message includes:
  - Specific reason: "Cannot connect to mcp-chrome-bridge"
  - Expected endpoint: "http://127.0.0.1:12306/mcp"
  - Troubleshooting steps:
    - "1. Start mcp-chrome-bridge: mcp-chrome-bridge"
    - "2. Verify connection: cf mcp-ping"
    - "3. Check Chrome extension is enabled"
  - Reference to troubleshooting guide

#### Scenario: Error message includes verification command

- **GIVEN** browser mode initialization failed
- **WHEN** error message is displayed
- **THEN** the message includes verification command: `cf mcp-ping`
- **AND** explains what the command does: "This will test your mcp-chrome setup"
- **AND** user can copy-paste the command to diagnose issues

### Requirement: Browser Mode Verification

The system SHALL provide commands to verify browser mode installation and diagnose issues.

#### Scenario: Ping MCP server

- **GIVEN** user wants to verify mcp-chrome setup
- **WHEN** user runs `cf mcp-ping`
- **THEN** the system attempts to connect to mcp-chrome
- **AND** if successful, displays: "✅ MCP Chrome Server is running on http://127.0.0.1:12306/mcp"
- **AND** if failed, displays: "❌ Cannot connect to MCP Chrome Server"
- **AND** the error message includes troubleshooting steps

#### Scenario: Test browser automation

- **GIVEN** user has verified mcp-chrome is running
- **WHEN** user runs `cf mocka`
- **THEN** the system opens Chrome browser
- **AND** navigates to Codeforces website
- **AND** demonstrates browser automation is working
- **AND** displays success or error message

#### Scenario: Test with actual command

- **GIVEN** user wants to test end-to-end functionality
- **WHEN** user runs `cf list <contest-id>` (e.g., `cf list 2122`)
- **THEN** the command should execute successfully
- **AND** display problem statistics table
- **AND** no Cloudflare errors occur
- **AND** if it fails, error message guides user to fix setup

### Requirement: Fetcher Interface

The system SHALL use a Fetcher interface that abstracts HTTP operations. BrowserFetcher implements this interface using mcp-chrome. HTTPFetcher is deprecated and removed from production use.

#### Scenario: Fetcher interface abstraction

- **GIVEN** the Fetcher interface defines:
  - `Get(url string) ([]byte, error)`
  - `GetJSON(url string) (map[string]interface{}, error)`
  - `Post(url string, data url.Values) ([]byte, error)`
- **WHEN** browser mode is enabled
- **THEN** client uses BrowserFetcher implementation
- **AND** BrowserFetcher uses mcp-chrome for all operations
- **AND** no code changes needed in client methods (Parse, Statis, etc.)

#### Scenario: BrowserFetcher GET request

- **GIVEN** browser mode is enabled
- **WHEN** code calls `fetcher.Get(url)`
- **THEN** BrowserFetcher uses mcp-chrome's GetWebContentHTML
- **AND** returns raw HTML from the URL
- **AND** bypasses Cloudflare protection
- **AND** includes cookies and session data from Chrome browser

#### Scenario: BrowserFetcher JSON request

- **GIVEN** browser mode is enabled
- **WHEN** code calls `fetcher.GetJSON(url)`
- **THEN** BrowserFetcher uses mcp-chrome's GetWebContent
- **AND** parses JSON from response
- **AND** returns parsed JSON object
- **AND** bypasses any anti-scraping measures

#### Scenario: BrowserFetcher POST request

- **GIVEN** browser mode is enabled
- **WHEN** code calls `fetcher.Post(url, data)`
- **THEN** BrowserFetcher uses mcp-chrome's NetworkRequest
- **AND** sends POST request with form data
- **AND** includes proper headers (Content-Type, etc.)
- **AND** returns response body

### Requirement: Command Categorization

The system SHALL categorize commands into network-dependent (require browser mode) and local-only (no browser required).

#### Scenario: Network-dependent commands require browser mode

- **GIVEN** user attempts to run network-dependent command
- **AND** command is one of: list, parse, submit, watch, race, pull, clone, upgrade
- **WHEN** browser mode is unavailable
- **THEN** the command fails immediately
- **AND** displays browser mode required error
- **AND** does not attempt HTTP fallback

#### Scenario: Local-only commands work independently

- **GIVEN** user attempts to run local-only command
- **AND** command is one of: config, test, gen, open, stand, sid
- **WHEN** browser mode is unavailable
- **THEN** the command executes successfully
- **AND** no browser mode error is displayed
- **AND** command performs its local operation

#### Scenario: Upgrade command special handling

- **GIVEN** user runs `cf upgrade`
- **WHEN** fetching from GitHub releases (no Cloudflare)
- **THEN** the system may use HTTP client directly (GitHub API allows it)
- **OR** use browser mode for consistency
- **AND** provide clear error if upgrade fails
- **AND** error message includes manual download link from GitHub
