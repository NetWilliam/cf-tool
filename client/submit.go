package client

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/NetWilliam/cf-tool/pkg/logger"
	"github.com/NetWilliam/cf-tool/util"

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

	body, err := util.GetBody(c.client, URL)
	if err != nil {
		logger.Error("Failed to fetch submit page: %v", err)
		return
	}

	logger.Debug("Fetched submit page: size=%d bytes", len(body))

	handle, err := findHandle(body)
	if err != nil {
		logger.Warning("Not logged in while submitting")
		return
	}

	logger.Info("Current user: %s", handle)
	fmt.Printf("Current user: %v\n", handle)

	csrf, err := findCsrf(body)
	if err != nil {
		logger.Error("Failed to extract CSRF token: %v", err)
		return
	}

	logger.Debug("CSRF token: %s", csrf)

	submitData := url.Values{
		"csrf_token":            {csrf},
		"ftaa":                  {c.Ftaa},
		"bfaa":                  {c.Bfaa},
		"action":                {"submitSolutionFormSubmitted"},
		"submittedProblemIndex": {info.ProblemID},
		"programTypeId":         {langID},
		"contestId":             {info.ContestID},
		"source":                {source},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
		"sourceCodeConfirmed":   {"true"},
	}

	logger.Debug("Submitting with data: programTypeId=%s, contestId=%s, problemIndex=%s",
		langID, info.ContestID, info.ProblemID)

	body, err = util.PostBody(c.client, fmt.Sprintf("%v?csrf_token=%v", URL, csrf), submitData)
	if err != nil {
		logger.Error("Failed to submit: %v", err)
		return
	}

	logger.Debug("Submit response size: %d bytes", len(body))

	errMsg, err := findErrorMessage(body)
	if err == nil {
		logger.Error("Submit error: %s", errMsg)
		return errors.New(errMsg)
	}

	msg, err := findMessage(body)
	if err != nil {
		logger.Error("Submit failed: %v", err)
		return errors.New("Submit failed")
	}
	if !strings.Contains(msg, "submitted successfully") {
		logger.Error("Submit failed with message: %s", msg)
		return errors.New(msg)
	}

	logger.Info("Code submitted successfully")
	color.Green("Submitted")

	submissions, err := c.WatchSubmission(info, 1, true)
	if err != nil {
		logger.Error("Failed to watch submission: %v", err)
		return
	}

	info.SubmissionID = submissions[0].ParseID()
	c.Handle = handle
	c.LastSubmission = &info

	logger.Info("Submission saved: ID=%s, handle=%s", info.SubmissionID, handle)
	return c.save()
}
