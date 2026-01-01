## Context

cf-tool is a CLI tool for competitive programmers on Codeforces. Codeforces uses Cloudflare protection that blocks direct HTTP requests from automated tools. To bypass this, cf-tool now uses **browser mode** via the mcp-chrome extension, which automates Chrome browser using Chrome DevTools Protocol.

**Current State:**
- `parse`, `submit`, `gen` already use browser mode
- Other commands fall back to HTTP mode when browser unavailable
- Mixed architecture creates inconsistent behavior

**Constraints:**
- Cloudflare blocks most direct HTTP requests to Codeforces
- Browser automation via mcp-chrome successfully bypasses Cloudflare
- Users must install mcp-chrome and Chrome to use cf-tool

**Stakeholders:**
- Competitive programmers using cf-tool
- Developers maintaining cf-tool
- mcp-chrome extension developers

## Goals / Non-Goals

**Goals:**
- Make browser mode the **only** supported mode for network operations
- Ensure all cf commands work consistently with browser mode
- Provide clear error messages when browser mode is unavailable
- Update documentation to reflect browser mode requirement
- Maintain backward compatibility for local-only operations (config, test, gen)

**Non-Goals:**
- Supporting HTTP mode as fallback (explicitly removing it)
- Implementing browser mode for commands that don't need it (config, test, open, stand, sid)
- Changing the behavior of already-migrated commands (parse, submit)
- Adding new features beyond browser mode migration

## Decisions

### Decision 1: Remove HTTP Mode Fallback

**Choice**: Completely remove HTTP mode fallback; make browser mode mandatory.

**Rationale**:
- Cloudflare blocks HTTP requests anyway, so HTTP mode doesn't work reliably
- Simplifies codebase by removing dual-mode architecture
- Clear user expectations: either browser mode works or cf-tool doesn't work
- Easier to maintain and debug single code path

**Alternatives Considered**:
1. **Keep HTTP fallback** - Rejected because HTTP mode is broken by Cloudflare
2. **Make HTTP opt-in with flag** - Rejected as unnecessary complexity
3. **Support both modes** - Rejected due to maintenance burden and user confusion

### Decision 2: Fetcher Interface Architecture

**Choice**: Keep existing Fetcher interface (HTTPFetcher vs BrowserFetcher) but only use BrowserFetcher.

**Rationale**:
- Already implemented and working
- Clean separation of concerns
- Easy to swap implementations if needed in future
- Minimal code changes required

**Current Architecture**:
```go
type Fetcher interface {
    Get(url string) ([]byte, error)
    GetJSON(url string) (map[string]interface{}, error)
    Post(url string, data url.Values) ([]byte, error)
}

type BrowserFetcher struct {
    mcpClient *mcp.Client
}
```

### Decision 3: Local-Only Commands No Changes

**Choice**: Commands that don't make network requests (config, test, open, stand, sid) should work without browser mode.

**Rationale**:
- These commands operate locally (file I/O, opening browser URLs)
- Don't interact with Codeforces API or pages
- Testing and configuration shouldn't require browser mode
- Reduces user friction for common workflows

**Commands Analysis**:
- `cf config` - Reads/writes local config file → No browser needed
- `cf test` - Compiles code locally, runs tests → No browser needed
- `cf open` - Opens URL in default browser → No browser mode needed
- `cf stand` - Opens URL in default browser → No browser mode needed
- `cf sid` - Opens URL in default browser → No browser mode needed

### Decision 4: Error Handling Strategy

**Choice**: Fail fast with clear error message at client initialization if browser unavailable.

**Rationale**:
- Prevents cryptic errors later in command execution
- Guides users immediately to the solution (install mcp-chrome)
- Consistent with "mandatory browser mode" decision

**Error Message Template**:
```
❌ Browser mode is required but not available.
Please install and configure mcp-chrome:
1. Download from: https://github.com/hangwin/mcp-chrome/releases
2. Run: npm install -g @hangwin/mcp-chrome-bridge
3. Start: mcp-chrome-bridge
4. Verify: cf mcp-ping

See README.md for detailed instructions.
```

## Implementation Details

### Client Initialization Flow

**Current (HTTP fallback)**:
```go
if err := c.initBrowserMode(); err != nil {
    logger.Warning("Browser mode not available: %v\n", err)
    logger.Info("Falling back to HTTP mode. Some features may not work.\n")
    c.fetcher = NewHTTPFetcher(c.client)  // FALLBACK
}
```

**New (mandatory browser)**:
```go
if err := c.initBrowserMode(); err != nil {
    return fmt.Errorf("browser mode is required: %w\n\n"+
        "Please install mcp-chrome:\n"+
        "  1. Download: https://github.com/hangwin/mcp-chrome/releases\n"+
        "  2. Run: npm install -g @hangwin/mcp-chrome-bridge\n"+
        "  3. Start: mcp-chrome-bridge\n"+
        "  4. Verify: cf mcp-ping\n\n"+
        "See README.md for details.", err)
}
c.fetcher = NewBrowserFetcher(c.mcpClient)
```

### Command Categorization

**Network-Dependent (require browser mode)**:
- `cf list` - Fetches problem statistics from Codeforces
- `cf parse` - Fetches problem page HTML (already uses browser)
- `cf submit` - Submits code via browser automation (already uses browser)
- `cf watch` - Fetches submission status from Codeforces
- `cf race` - Fetches contest info and countdown
- `cf pull` - Fetches submission source code
- `cf clone` - Fetches user's submission history
- `cf upgrade` - Fetches latest release from GitHub (can fallback to HTTP)

**Local-Only (no browser mode needed)**:
- `cf config` - Manages local configuration
- `cf gen` - Generates code from template (local)
- `cf test` - Compiles and tests locally
- `cf open` - Opens URL in default browser
- `cf stand` - Opens URL in default browser
- `cf sid` - Opens URL in default browser

## Risks / Trade-offs

### Risk 1: User Friction

**Risk**: All existing users must install mcp-chrome to continue using cf-tool.

**Impact**: High - breaking change for existing user base.

**Mitigation**:
- Clear error messages with installation instructions
- Update README with prominent warnings
- Provide migration guide in documentation
- Keep error messages helpful and actionable

### Risk 2: Browser Mode Failures

**Risk**: Browser mode may fail due to Chrome issues, extension issues, or network problems.

**Impact**: Medium - users unable to use cf-tool temporarily.

**Mitigation**:
- Robust error messages identifying specific failure points
- Troubleshooting guide in README
- Debug mode (`CF_DEBUG=1`) to help diagnose issues
- Verification commands (`cf mcp-ping`, `cf mocka`) to test setup

### Risk 3: Cross-Platform Compatibility

**Risk**: Browser mode may behave differently on different OSes (Linux, macOS, Windows).

**Impact**: Medium - inconsistent user experience.

**Mitigation**:
- Test on all supported platforms
- Document platform-specific issues
- Use standard Chrome DevTools Protocol (cross-platform)

### Risk 4: Performance Overhead

**Risk**: Browser automation is slower than direct HTTP requests.

**Impact**: Low - slight delay in command execution.

**Mitigation**:
- Acceptable trade-off for functionality (HTTP doesn't work anyway)
- Optimize browser automation where possible
- Cache results where appropriate

## Migration Plan

### Phase 1: Implementation (this change)

1. Modify client initialization to require browser mode
2. Test all network-dependent commands with browser mode
3. Verify local-only commands still work
4. Update documentation

### Phase 2: Release

1. Tag new version (e.g., v1.3.0)
2. Update CHANGELOG with breaking change notice
3. Release with prominent upgrade instructions

### Phase 3: User Migration

1. Users encounter error message pointing to mcp-chrome setup
2. Users follow documentation to install mcp-chrome
3. Users verify with `cf mcp-ping`
4. Users continue using cf-tool normally

### Rollback Plan

**If critical issues found**:
1. Revert HTTP fallback temporarily
2. Release patch version (e.g., v1.2.1)
3. Fix issues in branch
4. Re-apply mandatory browser mode in v1.3.0

**Note**: Rollback would only restore HTTP fallback as emergency measure. HTTP mode is broken by Cloudflare anyway, so this is just to buy time for fixes.

## Open Questions

### Question 1: GitHub Releases for `cf upgrade`

**Issue**: `cf upgrade` fetches from GitHub releases, which doesn't have Cloudflare protection.

**Options**:
1. Use browser mode for consistency
2. Allow HTTP fallback for GitHub only (simpler, faster)

**Recommendation**: Allow HTTP for GitHub (option 2). GitHub API doesn't require browser automation.

### Question 2: Graceful Degradation

**Issue**: Should we allow partial functionality if browser unavailable?

**Example**: User runs `cf test` (local-only) but browser not available.

**Options**:
1. Allow local commands to work even if browser unavailable
2. Fail entirely on any command if browser unavailable

**Recommendation**: Option 1 (graceful degradation). Local commands should work regardless of browser state.

### Question 3: Verification Commands

**Issue**: Should we add automated verification of browser setup on first run?

**Options**:
1. Run `cf mcp-ping` automatically on first run
2. Let user manually verify when they encounter issues

**Recommendation**: Option 2 (manual). Reduces startup time and doesn't surprise users.

## Testing Strategy

### Manual Testing Checklist

For each command:
1. Start with fresh session (rm ~/.cf/session)
2. Ensure mcp-chrome-bridge is running
3. Run command and verify success
4. Stop mcp-chrome-bridge
5. Run command and verify helpful error message

### Expected Results

**With mcp-chrome running**:
- All commands work successfully
- No Cloudflare errors
- No HTTP fallback warnings

**Without mcp-chrome running**:
- Network commands fail immediately with clear error
- Local commands (config, test, open, etc.) still work
- Error message guides user to install mcp-chrome

## Success Criteria

1. ✅ All network-dependent commands work with browser mode
2. ✅ All local-only commands work without browser mode
3. ✅ Clear error messages when browser unavailable
4. ✅ Documentation updated with browser mode requirements
5. ✅ No HTTP fallback code remains (or clearly marked as legacy)
6. ✅ Manual testing passes for all commands
