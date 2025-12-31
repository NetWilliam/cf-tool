# CF-Tool Login Flow Documentation

## Overview

CF-Tool supports two login modes:
1. **HTTP Login** - Traditional login with credentials (handle/password)
2. **Browser Mode Login** - Uses MCP Chrome Server for seamless authentication

## Login Flow Architecture

### Command Execution Flow

```
User runs: cf parse CONTEST 100
           ↓
    Parse() (cmd/parse.go)
           ↓
    ParseProblem() (client/parse.go)
           ↓
    fetcher.Get(URL) → body
           ↓
    findHandle(body) → ErrorNotLogged
           ↓
    loginAgain(cln, ErrorNotLogged)
           ↓
    Check password configuration
           ↓
    ┌────────────────────────┴────────────────────────┐
    │                                                 │
Password configured?                           No password
    │                                                 │
    ↓                                                 ↓
Login() with credentials                    tryInitBrowserClient()
    │                                                 │
    │                                                 ↓
    │                                    Detect MCP Server
    │                                                 │
    │                                                 ↓
    │                                    Create MCP Client
    │                                                 │
    │                                                 ↓
    │                                    SetMCPClient()
    │                                                 │
    │                                                 ↓
    └─────────────────────────────────────────────────┘
                        ↓
                  Login()
                        ↓
        ┌───────────────┴───────────────┐
        │                               │
browserEnabled=true?          browserEnabled=false
        │                               │
        ↓                               ↓
  LoginWithBrowser()              HTTP Login
```

## Key Components

### 1. Login Check (`findHandle`)

**Location:** `client/login.go`

```go
func findHandle(body []byte) (string, error) {
    reg := regexp.MustCompile(`handle = "([\s\S]+?)"`)
    tmp := reg.FindSubmatch(body)
    if len(tmp) < 2 {
        return "", errors.New(ErrorNotLogged)
    }
    return string(tmp[1]), nil
}
```

**Purpose:** Checks if user is logged in by searching for `handle` in HTML response.

### 2. Login Retry (`loginAgain`)

**Location:** `cmd/cmd.go`

```go
func loginAgain(cln *client.Client, err error) error {
    if err != nil && err.Error() == client.ErrorNotLogged {
        // Check if password is configured
        if len(cln.Password) == 0 || len(cln.HandleOrEmail) == 0 {
            // No password, try browser mode
            if browserErr := tryInitBrowserClient(cln); browserErr == nil {
                return cln.Login()  // Will use browser mode
            }
            // Browser mode unavailable, show configuration guide
        }
        // Password configured, normal login
        err = cln.Login()
    }
    return err
}
```

### 3. Browser Mode Auto-Detection (`tryInitBrowserClient`)

**Location:** `cmd/cmd.go`

```go
func tryInitBrowserClient(cln *client.Client) error {
    // Try to find MCP server
    serverURL, mcpPath, err := findMCPServer()
    if err != nil {
        return fmt.Errorf("MCP server not found: %w", err)
    }

    // Create MCP client (HTTP or stdio)
    var mcpClient *mcp.Client
    if serverURL != "" {
        mcpClient, err = mcp.NewClientHTTP(serverURL)
    } else if mcpPath != "" {
        mcpClient, err = mcp.NewClient("node", []string{mcpPath})
    }

    if err != nil {
        return err
    }

    // Enable browser mode
    cln.SetMCPClient(mcpClient)
    return nil
}
```

### 4. Login Method Selection (`Login`)

**Location:** `client/login.go`

```go
func (c *Client) Login() (err error) {
    // Check if browser mode is enabled
    if c.browserEnabled {
        return c.LoginWithBrowser()  // Browser mode
    }

    // Traditional HTTP login
    // ... requires password configuration
}
```

## Usage Scenarios

### Scenario 1: User with MCP Chrome Server (Recommended)

**Setup:**
```bash
# MCP Chrome Server is installed and running
# User has NOT configured password
```

**Execution:**
```bash
$ cf parse 100
Not logged. Try to login
No password configured. Attempting browser mode login...
Using HTTP transport: http://127.0.0.1:12306/mcp

✓ Already logged in as test_user
✓ Successfully parsed 3 samples
```

**What happens:**
1. Parse checks login → Not logged
2. System detects no password
3. Auto-detects MCP Chrome Server
4. Initializes browser mode
5. Opens browser for login (if not logged in)
6. Completes the original command

### Scenario 2: User with Password Configured

**Setup:**
```bash
$ cf config
handle/email: test_user
password: ********
```

**Execution:**
```bash
$ cf parse 100
Not logged. Try to login
Login test_user...
✓ Succeed!!
Welcome test_user~

✓ Successfully parsed 3 samples
```

**What happens:**
1. Parse checks login → Not logged
2. System detects password is configured
3. Uses HTTP login with credentials
4. Saves session
5. Completes the original command

### Scenario 3: No Credentials, No MCP Server

**Execution:**
```bash
$ cf parse 100
Not logged. Try to login
No password configured. Attempting browser mode login...

❌ Browser mode is not available.

Please configure your handle and password:
  Option 1: cf config
  Option 2: Install MCP Chrome Server for browser mode
```

**What happens:**
1. Parse checks login → Not logged
2. System detects no password
3. Tries to detect MCP Chrome Server → Fails
4. Shows clear error message with options
5. Exits with helpful guidance

## Benefits

### 1. Automatic Browser Mode Detection
- No manual configuration needed
- Automatically detects and uses available MCP Chrome Server
- Seamless user experience

### 2. Intelligent Fallback
```
Password configured?
├─ Yes → Use HTTP login
└─ No  → Try browser mode
         ├─ Success → Use browser mode
         └─ Fail → Show configuration guide
```

### 3. Better Error Messages
- Clear guidance on what to do
- Multiple options presented
- No confusion about next steps

### 4. Backward Compatibility
- Existing password-based login still works
- No breaking changes for current users
- Browser mode is opt-in enhancement

## Configuration Files

Browser mode settings (optional for auto-detection):

**`~/.cf/config`**
```json
{
  "browser": {
    "enabled": true,
    "mcp_transport": "http",
    "mcp_url": "http://127.0.0.1:12306/mcp"
  }
}
```

**Note:** If browser mode is not explicitly configured but MCP Chrome Server is detected, it will be used automatically when credentials are missing.

## Session Persistence

After successful login (HTTP or Browser):
- Session data saved to `~/.cf/session`
- `browserEnabled` flag saved (if browser mode used)
- Subsequent commands use saved session
- No need to login again unless session expires

## Troubleshooting

### "Not logged" Error

**Cause:** Session expired or not present

**Solutions:**
1. **Automatic:** Let the tool handle login (with browser mode or credentials)
2. **Manual:** Run `cf login` to explicitly login
3. **Configure:** Run `cf config` to set up credentials

### Browser Mode Not Working

**Check:**
```bash
$ cf mcp-ping
```

**Expected output:**
```
✓ MCP server is running!
Available tools:
  • chrome_navigate: Navigate to a URL
  • chrome_get_web_content: Get page content
  ...
```

**If failing:**
1. Ensure Chrome is running
2. Ensure MCP Chrome Extension is installed
3. Check MCP Native Host is configured
4. See `docs/LOGGING.md` for debugging with `logger.SetLevel(logger.DebugLevel)`

### Password Encryption

CF-Tool encrypts your password before saving:

```go
encrypt(handle, password) → encrypted_password
decrypt(handle, encrypted_password) → password
```

**Encryption:** AES-256-GCM with key derived from `glhf{handle}233`

**Storage:** `~/.cf/session`
```json
{
  "handle_or_email": "test_user",
  "password": "encrypted_hex_string_here",
  "browserEnabled": false
}
```

## Best Practices

### For Users

1. **Recommended Setup:** Install MCP Chrome Server for seamless authentication
2. **Alternative:** Configure password in `cf config`
3. **Debugging:** Enable debug logs if issues persist

### For Developers

1. **Always check login status** before operations requiring authentication
2. **Use `loginAgain()`** wrapper for automatic login retry
3. **Test both login modes** (HTTP and Browser)
4. **Handle ErrorNotLogged** gracefully

## Migration Guide

### From HTTP Login to Browser Mode

**Step 1:** Install MCP Chrome Server
```bash
# Follow instructions at:
# https://github.com/hangwin/mcp-chrome
```

**Step 2:** Test Installation
```bash
cf mcp-ping
```

**Step 3:** Use CF-Tool normally
```bash
cf parse 100  # Will auto-detect and use browser mode
```

**Step 4:** (Optional) Remove saved password
```bash
cf config
# Update password to empty string or remove it
```

## Code Examples

### Adding Login Check to New Commands

```go
func NewCommand() error {
    // Try operation first
    err := doOperation()

    // Automatically handle login
    if err != nil && err.Error() == client.ErrorNotLogged {
        if err = loginAgain(cln, err); err == nil {
            // Retry after successful login
            err = doOperation()
        }
    }

    return err
}
```

### Manual Login Check

```go
func checkLogin(cln *client.Client) bool {
    body, err := cln.fetcher.Get(cln.host)
    if err != nil {
        return false
    }

    _, err = findHandle(body)
    return err == nil
}
```
