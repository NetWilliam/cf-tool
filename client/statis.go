package client

import (
	"errors"
	"regexp"
	"strings"

	"github.com/NetWilliam/cf-tool/pkg/logger"
)

// StatisInfo statis information
type StatisInfo struct {
	ID     string
	Name   string
	IO     string
	Limit  string
	Passed string
	State  string
}

func findStatisBlock(body []byte) ([]byte, error) {
	logger.Debug("Finding statis block in HTML (size=%d bytes)", len(body))

	// Log a preview of the HTML content for debugging
	if len(body) > 0 {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		logger.Debug("HTML preview: %s", preview)
	}

	reg := regexp.MustCompile(`class="problems"[\s\S]+?</tr>([\s\S]+?)</table>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		logger.Error("Regex pattern did not match. Looking for 'class=\"problems\"' in HTML...")
		if !regexp.MustCompile(`class="problems"`).Match(body) {
			logger.Error("HTML does not contain 'class=\"problems\"' pattern")
		} else {
			logger.Info("Found 'class=\"problems\"' but full pattern didn't match")
		}
		return nil, errors.New("Cannot find any problem statis")
	}
	logger.Debug("Found statis block: %d bytes", len(tmp[1]))
	return tmp[1], nil
}

func findProblems(body []byte) ([]StatisInfo, error) {
	logger.Debug("Extracting problems from statis block (size=%d bytes)", len(body))

	reg := regexp.MustCompile(`<tr[\s\S]*?>`)
	tmp := reg.FindAllIndex(body, -1)
	if tmp == nil {
		logger.Error("Cannot find any problem rows")
		return nil, errors.New("Cannot find any problem")
	}

	logger.Debug("Found %d problem rows", len(tmp))

	ret := []StatisInfo{}
	scr := regexp.MustCompile(`<script[\s\S]*?>[\s\S]*?</script>`)
	cls := regexp.MustCompile(`class="(.+?)"`)
	rep := regexp.MustCompile(`<[\s\S]+?>`)
	ton := regexp.MustCompile(`<\s+`)
	rmv := regexp.MustCompile(`<+`)
	tmp = append(tmp, []int{len(body), 0})
	for i := 1; i < len(tmp); i++ {
		state := ""
		if x := cls.FindSubmatch(body[tmp[i-1][0]:tmp[i-1][1]]); x != nil {
			state = string(x[1])
		}
		b := scr.ReplaceAll(body[tmp[i-1][0]:tmp[i][0]], []byte{})
		b = rep.ReplaceAll(b, []byte("<"))
		b = ton.ReplaceAll(b, []byte("<"))
		b = rmv.ReplaceAll(b, []byte("<"))
		data := strings.Split(string(b), "<")
		tot := []string{}
		for j := 0; j < len(data); j++ {
			s := strings.TrimSpace(data[j])
			if s != "" {
				tot = append(tot, s)
			}
		}
		if len(tot) >= 5 {
			tot[4] = strings.ReplaceAll(tot[4], "x", "")
			tot[4] = strings.ReplaceAll(tot[4], "&nbsp;", "")
			if tot[4] == "" {
				tot[4] = "0"
			}
			ret = append(ret, StatisInfo{
				tot[0], tot[1], tot[2], tot[3],
				tot[4], state,
			})
			logger.Debug("Parsed problem: ID=%s, Name=%s, State=%s", tot[0], tot[1], state)
		}
	}

	logger.Info("Successfully parsed %d problems", len(ret))
	return ret, nil
}

// Statis get statis
func (c *Client) Statis(info Info) (problems []StatisInfo, err error) {
	URL, err := info.ProblemSetURL(c.host)
	if err != nil {
		logger.Error("Failed to build ProblemSetURL: %v", err)
		return
	}

	logger.Info("Fetching problem list from: %s", URL)

	if info.ProblemType == "acmsguru" {
		return nil, errors.New(ErrorNotSupportAcmsguru)
	}

	body, err := c.fetcher.Get(URL)
	if err != nil {
		logger.Error("Failed to fetch page: %v", err)
		return
	}

	logger.Debug("Fetched page: %d bytes", len(body))

	block, err := findStatisBlock(body)
	if err != nil {
		return
	}

	return findProblems(block)
}
