package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
)

// SubmitWithBrowser submits code using browser automation
func (c *Client) SubmitWithBrowser(info Info, langID, source string) (err error) {
	color.Cyan("Submit " + info.Hint() + " (Browser Mode)\n")

	if c.mcpClient == nil {
		return fmt.Errorf("MCP client not initialized. Please enable browser mode in config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Step 1: Navigate to submit page
	URL, err := info.SubmitURL(c.host)
	if err != nil {
		return err
	}

	color.Cyan("Navigating to submit page: %s", URL)
	if err := c.mcpClient.Navigate(ctx, URL); err != nil {
		return fmt.Errorf("failed to navigate to submit page: %w", err)
	}

	// Wait for page to load
	time.Sleep(2 * time.Second)

	// Step 2: Get page content and extract CSRF token
	color.Cyan("Extracting CSRF token...")
	content, err := c.mcpClient.GetWebContent(ctx, URL)
	if err != nil {
		return fmt.Errorf("failed to get page content: %w", err)
	}

	csrf, err := findCsrf([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to extract CSRF token: %w", err)
	}
	color.Green("✓ CSRF token extracted\n")

	// Get current handle
	handle, err := findHandle([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to get current handle: %w", err)
	}
	fmt.Printf("Current user: %v\n\n", handle)

	// Step 3: Inject code and submit using JavaScript
	color.Cyan("Submitting code using browser automation...")

	// Escape the source code for JavaScript
	escapedSource := jsEscapeString(source)

	// Build JavaScript to fill form and submit
	jsScript := fmt.Sprintf(`
		(function() {
			try {
				// Find and fill program type
				var programTypeSelect = document.querySelector('select[name="programTypeId"]');
				if (programTypeSelect) {
					programTypeSelect.value = "%s";
				}

				// Find and fill source code
				var sourceTextarea = document.querySelector('textarea[name="source"]');
				if (sourceTextarea) {
					sourceTextarea.value = %s;
				}

				// Find and fill other required fields
				var csrfInput = document.querySelector('input[name="csrf_token"]');
				if (csrfInput) {
					csrfInput.value = "%s";
				}

				var ftaaInput = document.querySelector('input[name="ftaa"]');
				if (ftaaInput) {
					ftaaInput.value = "%s";
				}

				var bfaaInput = document.querySelector('input[name="bfaa"]');
				if (bfaaInput) {
					bfaaInput.value = "%s";
				}

				// Submit the form
				var submitForm = document.querySelector('form[action*="submit"]');
				if (submitForm) {
					submitForm.submit();
					return {success: true, message: "Form submitted"};
				}

				// Alternative: click the submit button
				var submitButton = document.querySelector('input[type="submit"]');
				if (submitButton) {
					submitButton.click();
					return {success: true, message: "Submit button clicked"};
				}

				return {success: false, message: "Could not find submit form or button"};

			} catch(e) {
				return {success: false, message: e.toString()};
			}
		})();
	`, langID, escapedSource, csrf, c.Ftaa, c.Bfaa)

	// Execute JavaScript using chrome_javascript tool
	result, err := c.mcpClient.CallTool(ctx, "chrome_javascript", map[string]interface{}{
		"code": jsScript,
	})

	if err != nil {
		return fmt.Errorf("failed to execute JavaScript: %w", err)
	}

	if result.IsError {
		return fmt.Errorf("JavaScript execution failed")
	}

	// Parse result
	if len(result.Content) > 0 {
		if resultMap, ok := result.Content[0].(map[string]interface{}); ok {
			if resultText, ok := resultMap["text"].(string); ok {
				var jsResult map[string]interface{}
				if err := json.Unmarshal([]byte(resultText), &jsResult); err == nil {
					if success, ok := jsResult["success"].(bool); ok && !success {
						if msg, ok := jsResult["message"].(string); ok {
							return fmt.Errorf("JavaScript error: %s", msg)
						}
					}
				}
			}
		}
	}

	color.Green("✓ Code submitted successfully\n")

	// Wait for submission to process
	time.Sleep(2 * time.Second)

	// Step 4: Watch submission status
	color.Cyan("Watching submission status...")
	submissions, err := c.WatchSubmission(info, 1, true)
	if err != nil {
		return fmt.Errorf("failed to watch submission: %w", err)
	}

	color.Green("\n✓ Submission completed!\n")

	// Save submission info
	info.SubmissionID = submissions[0].ParseID()
	c.Handle = handle
	c.LastSubmission = &info

	return c.save()
}

// jsEscapeString escapes a string for use in JavaScript
func jsEscapeString(s string) string {
	// Use JSON encoding to properly escape the string
	b, _ := json.Marshal(s)
	return string(b)
}
