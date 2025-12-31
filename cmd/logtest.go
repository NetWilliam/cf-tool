package cmd

import (
	"fmt"
	"strings"

	"github.com/NetWilliam/cf-tool/pkg/logger"
	"github.com/fatih/color"
)

// LogTest demonstrates the logging system
func LogTest() error {
	color.Cyan("\n📝 CF-Tool Logging System Test\n")
	fmt.Println("========================================")

	// Show current log level
	currentLevel := logger.GetLevel()
	color.White("Current log level: %v\n", currentLevel)

	// Test different log levels (should not show as default is WARNING)
	logger.Debug("This DEBUG message should NOT be visible (default level is WARNING)")
	logger.Info("This INFO message should NOT be visible (default level is WARNING)")

	// These should be visible
	logger.Warning("This WARNING message SHOULD be visible")
	logger.Error("This ERROR message SHOULD be visible")

	fmt.Println("\n" + strings.Repeat("=", 50))
	color.Cyan("\n🔧 Setting log level to DEBUG to show all messages\n")
	fmt.Println("========================================")

	// Set to debug level
	logger.SetLevel(logger.DebugLevel)
	color.White("Log level set to: %v\n", logger.DebugLevel)

	// Now test all levels
	fmt.Println("\n--- Testing all log levels ---")
	logger.Debug("✓ Debug message is now visible")
	logger.Info("✓ Info message is now visible")
	logger.Warning("✓ Warning message")
	logger.Error("✓ Error message")

	// Test JSON logging
	fmt.Println("\n--- Testing JSON logging ---")

	testData := map[string]interface{}{
		"contestId":    "1234",
		"problemId":    "A",
		"submissionId": 5678,
		"verdict":      "Accepted",
		"passedTests":  10,
		"timeLimit":    "2000 ms",
		"memoryLimit":  "256 MB",
	}

	logger.DebugJSON("Submission Data", testData)

	// Test nested JSON
	nestedData := map[string]interface{}{
		"user": map[string]interface{}{
			"handle": "test_user",
			"rating": 1500,
			"rank":   "specialist",
		},
		"submissions": []interface{}{
			map[string]interface{}{"id": 1, "problem": "A", "verdict": "OK"},
			map[string]interface{}{"id": 2, "problem": "B", "verdict": "WA"},
		},
	}

	logger.DebugJSON("User Statistics", nestedData)

	// Test with large JSON
	largeData := map[string]interface{}{
		"problems": []map[string]interface{}{
			{"index": "A", "name": "Problem A", "score": 500},
			{"index": "B", "name": "Problem B", "score": 1000},
			{"index": "C", "name": "Problem C", "score": 1500},
			{"index": "D", "name": "Problem D", "score": 2000},
		},
		"contest": map[string]interface{}{
			"name": "Test Contest",
			"id":   1234,
			"type": "CF",
		},
	}

	logger.InfoJSON("Contest Data", largeData)

	fmt.Println("\n" + strings.Repeat("=", 50))
	color.Cyan("\n✅ Logging system test completed!\n")

	color.White("To enable debug logging in your code:")
	color.Cyan("  import \"github.com/NetWilliam/cf-tool/pkg/logger\"")
	color.Cyan("  logger.SetLevel(logger.DebugLevel)\n")

	color.White("Available log levels:")
	color.Cyan("  - logger.DebugLevel   // Detailed debugging info")
	color.Cyan("  - logger.InfoLevel    // General informational messages")
	color.Cyan("  - logger.WarningLevel // Warning messages (default)")
	color.Cyan("  - logger.ErrorLevel   // Error messages\n")

	return nil
}
