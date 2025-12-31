package client

import (
	"context"
	"errors"
	"time"

	"github.com/NetWilliam/cf-tool/client/browser"
	"github.com/NetWilliam/cf-tool/pkg/logger"

	"github.com/fatih/color"
)

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

	// Use browser automation to submit
	if err := browser.SubmitCode(ctx, c.mcpClient, URL, langID, source); err != nil {
		logger.Error("Failed to submit: %v", err)
		return err
	}

	logger.Info("Code submitted successfully")
	color.Green("Submitted")

	// Watch submission status
	// Add delay to ensure submission appears in the list
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
