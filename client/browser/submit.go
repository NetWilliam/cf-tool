package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NetWilliam/cf-tool/pkg/logger"
	"github.com/NetWilliam/cf-tool/pkg/mcp"
)

// SubmitCode performs browser automation to submit code
func SubmitCode(ctx context.Context, mcpClient *mcp.Client, URL, langID, source, problemID string) error {
	if mcpClient == nil {
		return errors.New("browser mode required")
	}

	logger.Info("Navigating to submit page: %s", URL)

	// Step 1: Navigate to submit page
	if err := mcpClient.Navigate(ctx, URL); err != nil {
		return fmt.Errorf("navigation failed: %w", err)
	}

	time.Sleep(2 * time.Second)

	// Step 2: Select problem (A/B/C/D/E) - CRITICAL FIX
	logger.Debug("Selecting problem: %s", problemID)
	if err := mcpClient.Fill(ctx, "[name='submittedProblemIndex']", problemID); err != nil {
		logger.Warning("Failed to fill problem selector: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Step 3: Select language
	logger.Debug("Selecting language: %s", langID)
	if err := mcpClient.Fill(ctx, "#programTypeId", langID); err != nil {
		logger.Warning("Failed to fill language selector with #programTypeId: %v", err)
		// Try alternative selector
		if err := mcpClient.Fill(ctx, "[name='programTypeId']", langID); err != nil {
			logger.Warning("Failed to fill language selector with [name='programTypeId']: %v", err)
		}
	}

	time.Sleep(500 * time.Millisecond)

	// Step 4: Inject source code using JavaScript
	logger.Debug("Injecting source code (%d bytes)...", len(source))

	// Escape source for JavaScript
	sourceEscaped := jsonEscape(source)

	jsCode := fmt.Sprintf(`
		(function() {
			let sourceField = document.querySelector('[name="source"]');
			if (!sourceField) {
				sourceField = document.getElementById('source');
			}
			if (sourceField) {
				sourceField.value = %s;
				return 'success';
			}
			return 'failed';
		})();
	`, sourceEscaped)

	_, err := mcpClient.CallTool(ctx, "chrome_javascript", map[string]interface{}{
		"code": jsCode,
	})
	if err != nil {
		return fmt.Errorf("failed to inject source code: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Step 5: Click submit button (use correct ID for Codeforces)
	logger.Debug("Clicking submit button...")
	submitSelectors := []string{
		"#singlePageSubmitButton",  // ✅ Codeforces specific button ID
		"input[type='submit']",
		"button[type='submit']",
		".submit",
		"[value='Submit']",
	}

	var submitErr error
	for _, selector := range submitSelectors {
		if err := mcpClient.Click(ctx, selector); err != nil {
			submitErr = err
			continue
		}
		submitErr = nil
		logger.Debug("Successfully clicked submit button with selector: %s", selector)
		break
	}

	if submitErr != nil {
		return fmt.Errorf("failed to click submit button: %w", submitErr)
	}

	// Wait for submission to process
	time.Sleep(3 * time.Second)

	logger.Info("Code submitted successfully via browser")
	return nil
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
