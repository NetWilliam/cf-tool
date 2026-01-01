# Change: Migrate Remaining Commands to Browser Mode

## Why

cf-tool has introduced browser mode to bypass Cloudflare protection on Codeforces. While `parse`, `submit`, and `gen` commands already use browser mode, the remaining commands still rely on the legacy HTTP fetcher. This creates inconsistency and potential failures when Cloudflare blocks direct HTTP requests.

**Problem**: Mixed architecture where some commands work via browser automation while others fall back to HTTP mode, leading to unpredictable behavior.

**Solution**: Make browser mode mandatory for all commands that interact with Codeforces, ensuring consistent Cloudflare bypass behavior across the entire tool.

## What Changes

- **BREAKING**: Remove HTTP mode fallback; browser mode (mcp-chrome) becomes mandatory for all network operations
- Update `cf list` to require browser mode (uses BrowserFetcher for problem statistics)
- Update `cf watch` to require browser mode (uses BrowserFetcher for submission monitoring)
- Update `cf race` to require browser mode (uses BrowserFetcher for contest countdown and parsing)
- Update `cf pull` to require browser mode (uses BrowserFetcher for fetching submissions)
- Update `cf clone` to require browser mode (uses BrowserFetcher for user submission history)
- Update `cf upgrade` to use browser mode for GitHub releases (fallback to HTTP if browser unavailable)
- Verify and fix `cf config` (local-only, should work without changes)
- Verify and fix `cf test` (local-only, should work without changes)
- Verify and fix `cf open` (opens URLs in browser, should work without changes)
- Verify and fix `cf stand` (opens URLs in browser, should work without changes)
- Verify and fix `cf sid` (opens URLs in browser, should work without changes)
- Add clear error messages when browser mode is unavailable
- Update documentation to reflect browser mode requirement

## Impact

- **Affected specs**:
  - `cli-display` (cf list command requires browser mode)
  - New spec needed: `browser-mode` (cross-cutting capability)
- **Affected code**:
  - `client/client.go` - Remove HTTP fallback, always require browser mode
  - `cmd/*.go` - Add browser mode checks where missing
  - `client/*.go` - Ensure all client methods work with BrowserFetcher
  - `README.md` - Update installation and usage documentation
- **Breaking change**: Users must install and configure mcp-chrome before using cf-tool
- **Migration impact**: Existing users will need to set up mcp-chrome to continue using cf-tool
