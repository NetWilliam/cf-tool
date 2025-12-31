package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/NetWilliam/cf-tool/pkg/logger"

	"github.com/fatih/color"
)

func findErrorMessage(body []byte) (string, error) {
	reg := regexp.MustCompile(`error[a-zA-Z_\-\ ]*">(.*?)</span>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return "", errors.New("Cannot find error")
	}
	return string(tmp[1]), nil
}

// Submit submit (block while pending)
func (c *Client) Submit(info Info, langID, source string) (err error) {
	color.Cyan("Submit " + info.Hint())

	logger.Info("Submitting code: problem=%s, langID=%s, sourceSize=%d bytes",
		info.Hint(), langID, len(source))

	URL, err := info.SubmitURL(c.host)
	if err != nil {
		logger.Error("Failed to build submit URL: %v", err)
		return
	}

	logger.Debug("Submit URL: %s", URL)

	// Check if we have browser mode available
	if !c.browserEnabled || c.mcpClient == nil {
		return errors.New("Browser mode is required for submit. Please ensure MCP Chrome Server is running.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Step 1: Navigate to submit page
	logger.Info("Navigating to submit page...")
	if err := c.mcpClient.Navigate(ctx, URL); err != nil {
		logger.Error("Failed to navigate to submit page: %v", err)
		return fmt.Errorf("navigation failed: %w", err)
	}

	// Wait a bit for page to load
	time.Sleep(2 * time.Second)

	// Step 2: Fill language selector
	logger.Debug("Selecting language: %s", langID)
	if err := c.mcpClient.Fill(ctx, "#programTypeId", langID); err != nil {
		logger.Warning("Failed to fill language selector: %v", err)
		// Try alternative selector
		if err := c.mcpClient.Fill(ctx, "[name='programTypeId']", langID); err != nil {
			logger.Warning("Failed to fill language selector with alternative: %v", err)
		}
	}

	// Step 3: Inject source code using JavaScript
	logger.Debug("Injecting source code...")
	// Escape the source code for JavaScript
	sourceEscaped := strings.ReplaceAll(source, "\\", "\\\\")
	sourceEscaped = strings.ReplaceAll(sourceEscaped, "`", "\\`")
	sourceEscaped = strings.ReplaceAll(sourceEscaped, "$", "\\$")
	sourceEscaped = strings.ReplaceAll(sourceEscaped, "'", "\\'")

	// Use JavaScript to directly set the source code field
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
	`, jsonEscape(source))

	// Execute JavaScript to set source code
	_, err = c.mcpClient.CallTool(ctx, "chrome_javascript", map[string]interface{}{
		"code": jsCode,
	})
	if err != nil {
		logger.Error("Failed to inject source code: %v", err)
		return fmt.Errorf("failed to inject source code: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Step 4: Click submit button
	logger.Debug("Clicking submit button...")
	submitSelectors := []string{
		"input[type='submit']",
		"button[type='submit']",
		".submit",
		"[value='Submit']",
	}

	var submitErr error
	for _, selector := range submitSelectors {
		if err := c.mcpClient.Click(ctx, selector); err != nil {
			submitErr = err
			continue
		}
		submitErr = nil
		break
	}

	if submitErr != nil {
		logger.Error("Failed to click submit button: %v", submitErr)
		return fmt.Errorf("failed to submit: %w", submitErr)
	}

	// Wait for submission to process
	time.Sleep(3 * time.Second)

	logger.Info("Code submitted successfully")
	color.Green("Submitted")

	// Step 5: Watch submission status
	// Add a bit more delay to ensure submission appears in the list
	logger.Info("Waiting for submission to appear in submission list...")
	time.Sleep(2 * time.Second)

	submissions, err := c.WatchSubmission(info, 1, true)
	if err != nil {
		logger.Error("Failed to watch submission: %v", err)
		logger.Warning("Submit was successful, but monitoring failed. You can check the status manually.")
		// Don't return error - the submission was successful
		return nil
	}

	info.SubmissionID = submissions[0].ParseID()
	c.LastSubmission = &info

	logger.Info("Submission saved: ID=%s", info.SubmissionID)
	return c.save()
}

// jsonEscape escapes a string for use in JavaScript
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
