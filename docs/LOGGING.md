# CF-Tool Logging System Documentation

## Overview

CF-Tool now includes a comprehensive logging system to help with debugging and monitoring. The logger supports multiple severity levels, colored console output, and structured JSON logging.

## Default Behavior

**Default log level: WARNING**

By default, only `WARNING` and `ERROR` level messages are displayed. This keeps the console output clean while still alerting you to potential issues.

## Log Levels

The logger supports four severity levels (in order of increasing severity):

1. **DEBUG** - Detailed diagnostic information for debugging
2. **INFO** - General informational messages about normal operation
3. **WARNING** - Warning messages for potentially harmful situations
4. **ERROR** - Error messages for critical issues

## Usage

### Basic Logging

```go
import "github.com/NetWilliam/cf-tool/pkg/logger"

// Debug level - only visible when log level is set to DEBUG
logger.Debug("Detailed debug info: variable=%v", variable)

// Info level - visible when log level is INFO or lower
logger.Info("Operation completed successfully")

// Warning level - always visible by default
logger.Warning("Potential issue detected: %s", warningMsg)

// Error level - always visible by default
logger.Error("Operation failed: %v", err)
```

### JSON Logging

For structured data, use the JSON logging functions:

```go
// Debug level JSON logging
data := map[string]interface{}{
    "contestId": "1234",
    "problemId": "A",
    "status": "Accepted",
}
logger.DebugJSON("Submission Data", data)

// Info level JSON logging
logger.InfoJSON("Response Data", response)
```

JSON data is automatically pretty-printed with indentation for better readability.

### Setting Log Level

```go
import "github.com/NetWilliam/cf-tool/pkg/logger"

// Enable all logs including DEBUG
logger.SetLevel(logger.DebugLevel)

// Enable INFO and above (hide DEBUG)
logger.SetLevel(logger.InfoLevel)

// Only show WARNING and ERROR (default)
logger.SetLevel(logger.WarningLevel)

// Only show ERROR
logger.SetLevel(logger.ErrorLevel)
```

### Configuring Output

```go
// Write to a file instead of stderr
file, err := os.OpenFile("cf-tool.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
if err != nil {
    log.Fatal(err)
}
logger.SetOutput(file)

// Disable colored output
logger.SetColor(false)
```

## What Gets Logged

### Fetcher Module

All HTTP/Browser requests are logged at DEBUG level:
- Request URL and method
- Response size in bytes
- Success/failure status

JSON responses are logged with `DebugJSON()` for inspection:
- API responses from Codeforces
- Form submission data
- Submission details

### Parse Module

- Problem URL and path being parsed
- Number of samples extracted
- File write operations
- Login status

### Submit Module

- Submission details (problem, language, source size)
- CSRF token extraction
- Form data being submitted
- Submission response
- Final submission ID

### Watch Module

- Submissions being fetched
- Number of submissions found
- Individual submission details (ID, problem, status)

## Examples

### Enable Debug Logging

```go
package main

import (
    "github.com/NetWilliam/cf-tool/pkg/logger"
)

func main() {
    // Enable debug logging at startup
    logger.SetLevel(logger.DebugLevel)

    // Now all debug logs will be visible
    // ... rest of your code
}
```

### Debugging Network Issues

When debugging network or browser issues, enable DEBUG level to see:
- All HTTP requests being made
- Request/response sizes
- JSON data being sent/received
- Browser automation details

```bash
# Run with debug logging (requires code change to set level)
cf submit 1234 A main.cpp
# Output will show all request details if logger.SetLevel(logger.DebugLevel) is called
```

## Testing the Logging System

Run the built-in logging test:

```bash
./bin/cf logtest
```

This will demonstrate:
- Current log level display
- Filtering behavior at different levels
- JSON pretty-printing
- Color-coded output

## Log Output Format

### Standard Messages

```
[DEBUG] Message here
[INFO ] Message here
[WARN ] Message here
[ERROR] Message here
```

### JSON Messages

```
[DEBUG] Submission Data:
{
  "contestId": "1234",
  "problemId": "A",
  "submissionId": 5678,
  "verdict": "Accepted"
}
```

### Colored Output

When enabled (default), log levels are color-coded:
- **DEBUG**: Cyan
- **INFO**: Green
- **WARNING**: Yellow
- **ERROR**: Red

## Performance Considerations

- Debug level logging has minimal performance impact when disabled
- JSON marshaling only occurs if the log level is enabled
- File I/O for log output is buffered

## Thread Safety

The logger is thread-safe and can be used concurrently from multiple goroutines.

## Best Practices

1. **Use appropriate log levels**:
   - DEBUG: Detailed diagnostics, variable values, step-by-step progress
   - INFO: High-level operation milestones
   - WARNING: Recoverable issues, deprecated usage
   - ERROR: Critical failures that prevent operation

2. **Log structured data with JSON functions**:
   ```go
   logger.DebugJSON("API Response", response)  // Good
   logger.Debug(fmt.Sprintf("Response: %+v", response))  // Less readable
   ```

3. **Include context in log messages**:
   ```go
   logger.Error("Failed to parse problem %s: %v", problemID, err)  // Good
   logger.Error("Parse failed: %v", err)  // Less context
   ```

4. **Don't log sensitive data**:
   - Avoid logging passwords, tokens, or private keys
   - Be careful with user data

5. **Use DEBUG for development, WARNING for production**:
   ```go
   if os.Getenv("CF_DEBUG") != "" {
       logger.SetLevel(logger.DebugLevel)
   }
   ```
