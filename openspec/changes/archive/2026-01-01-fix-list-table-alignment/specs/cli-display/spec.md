## ADDED Requirements

### Requirement: Contest Problem List Table Display

The system SHALL display contest problem statistics in a well-formatted table when users execute the `cf list` command. The table MUST include columns for problem identifier, problem name, number of passed submissions, time/memory limits, and input/output specifications.

#### Scenario: Display problem list with aligned columns

- **GIVEN** a valid contest identifier (e.g., 2122, 2050)
- **WHEN** the user runs `cf list <contest-id>`
- **THEN** the system displays a table with:
  - Column "#": Problem identifier (A, B, C, etc.), left-aligned
  - Column "PROBLEM": Problem name, left-aligned, auto-wrapped if too long
  - Column "PASSED": Number of passed submissions, right-aligned
  - Column "LIMIT": Time and memory limits (e.g., "1 s, 256 MB"), left-aligned
  - Column "IO": Input/output specifications, left-aligned
- **AND** all column borders align properly to form a readable grid
- **AND** the table fits within standard terminal width (80-120 characters)
- **AND** numeric columns (PASSED) are right-aligned for easy comparison

#### Scenario: Handle varying problem name lengths

- **GIVEN** a contest with short problem names (e.g., "Greedy Grid")
- **AND** a contest with long problem names (e.g., "Digital string maximization")
- **WHEN** displaying the problem list table
- **THEN** short names display without excessive whitespace
- **AND** long names either wrap or truncate gracefully
- **AND** column alignment remains consistent across all rows

#### Scenario: Handle varying PASSED count magnitudes

- **GIVEN** problems with different submission counts:
  - Small: 124 submissions (3 digits)
  - Medium: 7,871 submissions (4 digits with comma)
  - Large: 18,561 submissions (5 digits with comma)
- **WHEN** displaying the problem list table
- **THEN** all PASSED values align on the right edge
- **AND** numbers are formatted consistently (with or without comma separators)
- **AND** the column width accommodates the largest value without overflow

#### Scenario: Handle mixed language IO text

- **GIVEN** problems with IO specifications in different languages:
  - English: "standard input/output"
  - Chinese: "标准输入/输出"
- **WHEN** displaying the problem list table
- **THEN** the IO column displays text correctly
- **AND** Unicode characters (Chinese, Cyrillic, etc.) render properly
- **AND** column alignment accounts for Unicode character width

#### Scenario: Maintain color highlighting for problem states

- **GIVEN** problems with different submission states (accepted, rejected)
- **WHEN** displaying the problem list table
- **THEN** accepted problem rows highlight with green background
- **AND** rejected problem rows highlight with red background
- **AND** color formatting does not break table alignment
- **AND** table borders and structure remain visible through color highlights

#### Scenario: Handle extreme terminal widths

- **GIVEN** a terminal with narrow width (80 characters)
- **WHEN** displaying a problem list with long names
- **THEN** the table either:
  - Wraps long text to fit within width
  - OR truncates with ellipsis (...) to maintain alignment
- **AND** all columns remain properly aligned
- **AND** the table remains readable without horizontal scrolling

### Requirement: Table Configuration

The system SHALL configure the tablewriter library with appropriate settings to ensure consistent table rendering across all contests and terminal environments.

#### Scenario: Configure column alignments

- **GIVEN** the tablewriter library initialization
- **WHEN** setting up the table for problem display
- **THEN** the system configures:
  - Column 0 (#): Left-aligned (text identifier)
  - Column 1 (PROBLEM): Left-aligned (variable length text)
  - Column 2 (PASSED): Right-aligned (numeric data)
  - Column 3 (LIMIT): Left-aligned (formatted text)
  - Column 4 (IO): Left-aligned (descriptive text)

#### Scenario: Configure column widths

- **GIVEN** the tablewriter library initialization
- **WHEN** setting up the table for problem display
- **THEN** the system sets:
  - Minimum column widths to prevent cramped display
  - Maximum column widths for PROBLEM column to prevent overflow
  - Auto-width calculation for other columns based on content
  - Reasonable default widths if content is unavailable

#### Scenario: Configure auto-wrap behavior

- **GIVEN** problem names that exceed column width
- **WHEN** rendering the table
- **THEN** the system:
  - Enables auto-wrap for the PROBLEM column
  - Wraps text at word boundaries when possible
  - Maintains row height consistency
  - Does not wrap other columns unnecessarily

#### Scenario: Configure table borders and padding

- **GIVEN** the tablewriter library initialization
- **WHEN** setting up the table for problem display
- **THEN** the system configures:
  - Clear border characters for table grid
  - Appropriate cell padding (1-2 spaces) for readability
  - Header separator line for distinction from data rows
  - Consistent border style across all table edges
