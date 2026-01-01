# Change: Fix cf List Table Alignment

## Why

The `cf list` command displays contest problem statistics in a table format, but users report alignment issues with certain contests (e.g., contest 2122) where columns do not align properly, particularly when:
- Problem names vary significantly in length
- PASSED column contains numbers with different digit counts (3-5 digits)
- IO column contains mixed English/Chinese text
- Terminal width is limited

The current implementation uses tablewriter with default settings, which does not configure column widths, alignments, or wrapping behavior, leading to inconsistent display across different contests and terminal sizes.

## What Changes

- Configure tablewriter to set proper column alignments (right-align numeric columns like PASSED)
- Set appropriate column widths or max widths to prevent overflow
- Enable auto-wrap for long problem names to maintain table within terminal bounds
- Improve numeric column formatting for consistent display
- Ensure table respects common terminal widths (80-120 characters)

## Impact

- Affected specs: `cli-display` (new capability for command-line output formatting)
- Affected code: `cmd/list.go` (table rendering logic)
- User-facing: Users will see properly aligned tables when running `cf list` for any contest
- Backward compatibility: Display format will be similar but with improved alignment
