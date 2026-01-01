# Codeforces Tool

[![Build Status](https://travis-ci.org/xalanq/cf-tool.svg?branch=master)](https://travis-ci.org/xalanq/cf-tool)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/cf-tool)](https://goreportcard.com/report/github.com/xalanq/cf-tool)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.12-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/cf-tool/master/LICENSE)

Codeforces Tool is a command-line interface tool for [Codeforces](https://codeforces.com).

This is a fork of the original [cf-tool](https://github.com/xalanq/cf-tool) with browser mode support to bypass Cloudflare protection.

[中文](./README_zh_CN.md)

## Installation

### Step 1: Build cf-tool

Clone the repository and build with make **(go >= 1.12)**:

```bash
# Clone the repository
git clone https://github.com/NetWilliam/cf-tool.git
cd cf-tool

# Build with make
make build

# (Optional) Install to ~/go/bin
make install
```

The compiled binary will be at `./bin/cf`. You can move it to anywhere you like or add it to your PATH.

### Step 2: Install Browser Mode Components

⚠️ **IMPORTANT**: Browser Mode is **REQUIRED** for all network operations (parse, submit, list, watch, race, pull, clone).

cf-tool uses **Browser Mode** to bypass Cloudflare protection on Codeforces. You need to install:

1. **[mcp-chrome](https://github.com/hangwin/mcp-chrome/)** - Chrome extension that exposes Chrome DevTools Protocol via MCP

2. **mcp-chrome-bridge** - Node.js bridge service

#### Quick Installation

```bash
# Install mcp-chrome-bridge
npm install -g @hangwin/mcp-chrome-bridge

# Or using pnpm
pnpm add -g @hangwin/mcp-chrome-bridge
```

Then:

1. Download [mcp-chrome extension](https://github.com/hangwin/mcp-chrome/releases)
2. Extract to a folder
3. Open Chrome and go to `chrome://extensions/`
4. Enable "Developer mode"
5. Click "Load unpacked" and select the extension folder
6. Run `mcp-chrome-bridge` in a terminal (it runs on `http://127.0.0.1:12306/mcp`)

#### Verify Installation

```bash
# Test MCP connection
cf mcp-ping

# Test browser automation
cf mocka
```

**Important**: Make sure both commands succeed before using cf-tool!

For more details about mcp-chrome, visit: https://github.com/hangwin/mcp-chrome/

## New Commands

This fork adds two new commands for browser mode testing:

### cf mcp-ping

Test the connection to MCP Chrome Server.

```bash
cf mcp-ping
```

Expected output:
```
✅ MCP Chrome Server is running
```

### cf mocka

Test browser automation capabilities. Opens Chrome, navigates to Google and searches for "billboard quarterly chart", then returns page content to verify browser automation is working correctly.

```bash
cf mocka
```

This command is used to verify that cf-tool can correctly control your browser.

## Verified Commands

The following commands have been tested and verified to work with browser mode:

- [x] `cf parse` - Fetch problem samples
- [x] `cf gen` - Generate code from template
- [x] `cf test` - Compile and test locally
- [x] `cf submit` - Submit code to Codeforces
- [x] `cf open` - Open problems in browser
- [x] `cf sid` - Open submission page
- [x] `cf race` - Contest countdown and parsing
- [ ] `cf clone` - Clone user submissions (~~partially working~~, behavior unclear)

## Original Documentation

For detailed usage of all cf-tool commands (parse, submit, race, etc.), please refer to the [original repository](https://github.com/xalanq/cf-tool).

All original commands are supported in this fork with browser mode enabled by default for network operations.

## FAQ

### Browser mode is required?

Yes. Due to Cloudflare protection on Codeforces, all network-dependent commands now require browser mode.

### How to check if browser mode is working?

Run `cf mcp-ping`. If it shows "✅ MCP Chrome Server is running", browser mode is ready.

### Can I use the old HTTP mode?

No. The old HTTP mode cannot bypass Cloudflare protection and is no longer supported.

## License

MIT License - Same as the original [cf-tool](https://github.com/xalanq/cf-tool)
