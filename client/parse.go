package client

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/NetWilliam/cf-tool/pkg/logger"

	"github.com/k0kubun/go-ansi"

	"github.com/fatih/color"
)

func findSample(body []byte) (input [][]byte, output [][]byte, err error) {
	logger.Debug("Finding samples in HTML (size=%d bytes)", len(body))

	// Find all <pre> tags that are inside input or output divs
	// This approach works regardless of nesting level

	// First, find all input blocks with their <pre> content
	inputReg := regexp.MustCompile(`<div[^>]*class="input"[^>]*>[\s\S]*?<pre[^>]*>([\s\S]*?)</pre>`)
	inputMatches := inputReg.FindAllSubmatch(body, -1)

	// Find all output blocks with their <pre> content
	outputReg := regexp.MustCompile(`<div[^>]*class="output"[^>]*>[\s\S]*?<pre[^>]*>([\s\S]*?)</pre>`)
	outputMatches := outputReg.FindAllSubmatch(body, -1)

	logger.Debug("Found %d input blocks and %d output blocks", len(inputMatches), len(outputMatches))

	if len(inputMatches) == 0 || len(outputMatches) == 0 {
		logger.Error("Cannot find any sample input/output blocks")
		return nil, nil, fmt.Errorf("Cannot parse sample with input %v and output %v", len(inputMatches), len(outputMatches))
	}

	// Ensure we have the same number of inputs and outputs
	count := len(inputMatches)
	if len(outputMatches) < count {
		count = len(outputMatches)
	}

	for i := 0; i < count; i++ {
		// Process input: remove HTML tags, unescape, clean whitespace
		inputContent := extractTextContent(inputMatches[i][1])
		input = append(input, []byte(inputContent+"\n"))

		// Process output: remove HTML tags, unescape, clean whitespace
		outputContent := extractTextContent(outputMatches[i][1])
		output = append(output, []byte(outputContent+"\n"))

		logger.Debug("Extracted sample %d: input=%d bytes, output=%d bytes", i+1, len(inputContent), len(outputContent))
	}

	logger.Debug("Found %d sample pairs", len(input))
	return
}

// extractTextContent extracts text content from HTML, removing all tags
func extractTextContent(htmlBytes []byte) string {
	// Remove all HTML tags
	tagReg := regexp.MustCompile(`<[^>]+>`)
	text := tagReg.ReplaceAllString(string(htmlBytes), "")

	// Unescape HTML entities
	text = html.UnescapeString(text)

	// Clean up whitespace: replace <br>, &nbsp;, etc already handled
	// Normalize whitespace
	spaceReg := regexp.MustCompile(`\s+`)
	text = spaceReg.ReplaceAllString(text, " ")

	// Trim and convert HTML line breaks to newlines
	text = strings.TrimSpace(text)

	return text
}

// ParseProblem parse problem to path. mu can be nil
func (c *Client) ParseProblem(URL, path string, mu *sync.Mutex) (samples int, standardIO bool, err error) {
	logger.Info("Parsing problem: URL=%s, path=%s", URL, path)

	body, err := c.fetcher.Get(URL)
	if err != nil {
		logger.Error("Failed to fetch problem page: %s - %v", URL, err)
		return
	}

	logger.Debug("Fetched problem page: size=%d bytes", len(body))

	input, output, err := findSample(body)
	if err != nil {
		logger.Error("Failed to extract samples: %v", err)
		return
	}

	logger.Info("Extracted %d sample(s)", len(input))

	standardIO = true
	if !bytes.Contains(body, []byte(`<div class="input-file"><div class="property-title">input</div>standard input</div><div class="output-file"><div class="property-title">output</div>standard output</div>`)) {
		standardIO = false
	}

	logger.Debug("Standard IO: %v", standardIO)

	for i := 0; i < len(input); i++ {
		fileIn := filepath.Join(path, fmt.Sprintf("in%v.txt", i+1))
		fileOut := filepath.Join(path, fmt.Sprintf("ans%v.txt", i+1))
		e := os.WriteFile(fileIn, input[i], 0644)
		if e != nil {
			if mu != nil {
				mu.Lock()
			}
			color.Red(e.Error())
			if mu != nil {
				mu.Unlock()
			}
			logger.Error("Failed to write input file %s: %v", fileIn, e)
		} else {
			logger.Debug("Wrote input file: %s (%d bytes)", fileIn, len(input[i]))
		}
		e = os.WriteFile(fileOut, output[i], 0644)
		if e != nil {
			if mu != nil {
				mu.Lock()
			}
			color.Red(e.Error())
			if mu != nil {
				mu.Unlock()
			}
			logger.Error("Failed to write output file %s: %v", fileOut, e)
		} else {
			logger.Debug("Wrote output file: %s (%d bytes)", fileOut, len(output[i]))
		}
	}

	logger.Info("Successfully parsed %d samples", len(input))
	return len(input), standardIO, nil
}

// Parse parse
func (c *Client) Parse(info Info) (problems []string, paths []string, err error) {
	color.Cyan("Parse " + info.Hint())

	logger.Debug("Parse info: ProblemID=%s, ProblemType=%s", info.ProblemID, info.ProblemType)

	problemID := info.ProblemID
	info.ProblemID = "%v"
	urlFormatter, err := info.ProblemURL(c.host)
	if err != nil {
		logger.Error("Failed to build ProblemURL: %v", err)
		return
	}
	info.ProblemID = ""

	logger.Debug("URL formatter: %s", urlFormatter)

	if problemID == "" {
		logger.Info("No problemID specified, fetching problem list from contest page...")
		statics, err := c.Statis(info)
		if err != nil {
			logger.Error("Failed to get problem statistics: %v", err)
			return nil, nil, err
		}
		logger.Info("Found %d problems in contest", len(statics))
		problems = make([]string, len(statics))
		for i, problem := range statics {
			problems[i] = problem.ID
		}
	} else {
		problems = []string{problemID}
	}
	contestPath := info.Path()
	logger.Info("The problem(s) will be saved to %v", contestPath)

	wg := sync.WaitGroup{}
	wg.Add(len(problems))
	mu := sync.Mutex{}
	paths = make([]string, len(problems))
	for i, problemID := range problems {
		paths[i] = filepath.Join(contestPath, strings.ToLower(problemID))
		go func(problemID, path string) {
			defer wg.Done()
			mu.Lock()
			fmt.Printf("Parsing %v\n", problemID)
			mu.Unlock()

			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return
			}
			URL := fmt.Sprintf(urlFormatter, problemID)

			samples, standardIO, err := c.ParseProblem(URL, path, &mu)
			if err != nil {
				return
			}

			warns := ""
			if !standardIO {
				warns = color.YellowString("Non standard input output format.")
			}
			mu.Lock()
			if err != nil {
				color.Red("Failed %v. Error: %v", problemID, err.Error())
			} else {
				ansi.Printf("%v %v\n", color.GreenString("Parsed %v with %v samples.", problemID, samples), warns)
			}
			mu.Unlock()
		}(problemID, paths[i])
	}
	wg.Wait()
	return
}
