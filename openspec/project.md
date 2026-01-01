# Project Context

## Purpose

**cf-tool** (Codeforces Tool) is a command-line interface tool designed to help competitive programmers participate in Codeforces contests more efficiently. The tool automates common workflows including:

- Submitting code solutions to Codeforces contests, Gym, Groups, and acmsguru
- Fetching problem statements and test cases
- Local compilation and testing of solutions before submission
- Real-time submission status monitoring
- Code template generation with customizable placeholders
- Cloning other users' submissions for learning
- Browser automation to bypass Cloudflare protection

The project's goal is to provide a fast, lightweight, cross-platform CLI tool that streamlines the competitive programming workflow on Codeforces.

## Tech Stack

### Primary Technologies
- **Go 1.24+** - Main programming language
- **Go Modules** - Dependency management
- **Make** - Build automation and cross-platform compilation

### Key Dependencies
- `github.com/PuerkitoBio/goquery` - HTML parsing for web scraping
- `github.com/docopt/docopt-go` - CLI argument parsing
- `github.com/fatih/color` - Colored console output
- `github.com/mitchellh/go-homedir` - Home directory path handling
- `github.com/olekukonko/tablewriter` - Table formatting for CLI
- `github.com/sergi/go-diff` - Text diffing utilities
- `github.com/shirou/gopsutil` - System utilities
- `golang.org/x/crypto` - Cryptographic functions

### External Services
- **mcp-chrome** - Chrome extension that exposes Chrome DevTools Protocol via MCP (required for browser automation)
- **mcp-chrome-bridge** - Node.js bridge service for MCP Chrome communication (runs on http://127.0.0.1:12306/mcp)
- **Codeforces** - Primary platform API and web interface

## Project Conventions

### Code Style

**Formatting:**
- Use standard `go fmt` formatting
- Follow Go conventions from [Effective Go](https://golang.org/doc/effective_go)
- Run `make fmt` before committing changes
- Use `go vet` to catch potential issues: `make vet`

**Naming Conventions:**
- Package names: lowercase, single words when possible (e.g., `client`, `config`, `util`)
- Exported functions: PascalCase (e.g., `ParseProblem`, `SubmitCode`)
- Internal functions: camelCase (e.g., `fetchHTML`, `parseSamples`)
- Constants: PascalCase or UPPER_CASE
- Interface names: usually -er suffix (e.g., `Fetcher`, `Submitter`)

**File Organization:**
- Main entry point: `cf.go` (not in a package directory)
- Command implementations: `cmd/*.go`
- Core logic: `client/*.go`
- Utilities: `util/*.go`
- Configuration: `config/*.go`
- Package libraries: `pkg/*/`

### Architecture Patterns

**Layered Architecture:**
```
├── cf.go              # Main entry point
├── cmd/               # CLI command layer (argument parsing, routing)
├── client/            # Business logic layer (API interaction, submission)
│   └── browser/       # Browser automation layer (Chrome DCP)
├── pkg/               # Shared packages (logger, MCP client, types)
├── config/            # Configuration management
├── cookiejar/         # Session management
└── util/              # Utility functions
```

**Key Patterns:**
- **Command Pattern**: Each CLI command is a separate function in `cmd/*.go`
- **Adapter Pattern**: `client/browser/adapter.go` adapts browser automation to the submission interface
- **Template Method**: Code generation uses configurable templates with placeholders
- **Strategy Pattern**: Multiple submission strategies (legacy API vs browser automation)

**Browser Mode Architecture:**
- Browser mode is now **required** for `parse` and `submit` commands
- Uses Chrome DevTools Protocol via MCP for web automation
- Bypasses Cloudflare protection that blocks direct API access
- Falls back gracefully when browser is unavailable

### Testing Strategy

**Current State:**
- Limited automated test coverage (no dedicated test files)
- Manual testing through CLI commands:
  - `cf mocka` - Test browser automation
  - `cf mcp-ping` - Test MCP Chrome connection
  - `cf logtest` - Test logging system
- Travis CI runs basic `go build` validation

**Testing Commands:**
- `make test` - Run Go tests (minimal coverage currently)
- `make test-coverage` - Generate coverage report
- `make check` - Run fmt, vet, and test

**Testing Guidelines:**
- Before submitting changes, run `make dev` to format, vet, and build
- Test browser automation with `cf mocka` after any MCP-related changes
- Verify MCP connection with `cf mcp-ping`
- Manual testing of the full workflow: `cf race`, `cf parse`, `cf test`, `cf submit`

### Git Workflow

**Branching:**
- `master` - Main development branch
- Feature branches - Create from `master` for new features
- Format: `feature/feature-name` or `fix/bug-description`

**Commit Conventions:**
- Follow conventional commit format: `type: description`
- Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`
- Examples:
  - `feat: add support for Codeforces Gym contests`
  - `fix: resolve Cloudflare challenge detection`
  - `docs: update installation instructions for mcp-chrome`
  - `refactor: modernize Go code to use new APIs`

**Commit Message Body:**
- Add detailed explanation for complex changes
- Reference issue numbers if applicable

**Pull Requests:**
- Ensure `make check` passes before PR
- Include test steps for manual verification
- Document any breaking changes

## Domain Context

**Codeforces Platform:**
- Competitive programming website (https://codeforces.com)
- Contest types: Regular contests, Gym (practice), Groups, acmsguru
- Problem format: Multiple test cases with input/output samples
- Submission workflow: Upload code → Run against tests → Get verdict

**Contest Types:**
- **Contests**: Regular timed competitions
- **Gym**: Practice contests without formal ranking
- **Groups**: Private or public custom contests
- **acmsguru**: Archive of problems from Timus Online Judge

**Submission Verdicts:**
- Accepted (AC)
- Wrong Answer (WA)
- Time Limit Exceeded (TLE)
- Memory Limit Exceeded (MLE)
- Runtime Error (RE)
- Compilation Error (CE)
- Hacked / Hacking

**Template Placeholders:**
- `$%U%$` - User handle
- `$%Y%$` - Year (e.g., 2025)
- `$%M%$` - Month (e.g., 01)
- `$%D%$` - Day (e.g., 01)
- `$%h%$` - Hour
- `$%m%$` - Minute
- `$%s%$` - Second

**File Structure:**
- `~/.cf/config` - User configuration (templates, defaults)
- `~/.cf/session` - Session data (cookies, credentials)
- `./cf/contest/<id>/<problem>` - Default problem directory

## Important Constraints

**Technical Constraints:**
- **Go 1.12+ required** for building (currently using Go 1.24)
- **Browser mode mandatory** for `parse` and `submit` commands (Cloudflare bypass)
- **Cross-platform support**: Linux, macOS, Windows (amd64/arm64)
- **Single binary distribution**: No external runtime dependencies except mcp-chrome

**Platform Constraints:**
- **Codeforces API limitations** - Rate limiting, Cloudflare protection
- **Browser requirement** - Must have Chrome/Chromium installed
- **mcp-chrome dependency** - External service must be running on port 12306

**Business/Usage Constraints:**
- Must not violate Codeforces terms of service
- Should respect rate limits and avoid excessive requests
- User credentials stored locally (encrypted in session file)
- Tool is for individual use, not for automated bulk operations

**Security Constraints:**
- Handle user credentials securely (password in config file)
- Don't log sensitive information (cookies, session tokens)
- Validate all user inputs before executing
- Use HTTPS for all network communications

## External Dependencies

### Required Services

**mcp-chrome Extension:**
- **Purpose**: Chrome DevTools Protocol bridge for browser automation
- **Repository**: https://github.com/hangwin/mcp-chrome/
- **Installation**: Download from releases, load as unpacked extension
- **Configuration**: Exposes MCP server on http://127.0.0.1:12306/mcp

**mcp-chrome-bridge:**
- **Purpose**: Node.js proxy service for MCP Chrome communication
- **Installation**: `npm install -g @hangwin/mcp-chrome-bridge`
- **Usage**: Run `mcp-chrome-bridge` before using cf-tool
- **Default endpoint**: http://127.0.0.1:12306/mcp

### External APIs

**Codeforces:**
- **Website**: https://codeforces.com
- **API**: Unofficial API (HTML scraping for most operations)
- **Authentication**: Cookie-based session management
- **Cloudflare**: Requires browser automation to bypass

### Build Dependencies

**Go Toolchain:**
- Go 1.24+ compiler
- Go modules for dependency management
- Standard library packages: `net/http`, `os/exec`, `io`, `time`, etc.

**Cross-Compilation:**
- Supports Linux (amd64, arm64)
- Supports macOS (amd64, arm64/Apple Silicon)
- Supports Windows (amd64, arm64)

### Development Tools

**Make Targets:**
- `make build` - Build for current platform
- `make install` - Install to GOPATH/bin
- `make dev` - Format, vet, and build
- `make check` - Run all checks
- `make build-all` - Build for all platforms

**CI/CD:**
- Travis CI for automated builds
- Runs `go build` on all pull requests
- No automated testing currently
