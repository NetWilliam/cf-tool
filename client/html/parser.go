package html

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/NetWilliam/cf-tool/pkg/logger"
)

// ParseTestcases extracts test cases from problem HTML
// Works with Codeforces problem page format
func ParseTestcases(body []byte) (input, output [][]byte, err error) {
	// Find all <pre> tags that are inside input or output divs
	inputReg := regexp.MustCompile(`<div[^>]*class="input"[^>]*>[\s\S]*?<pre[^>]*>([\s\S]*?)</pre>`)
	inputMatches := inputReg.FindAllSubmatch(body, -1)

	outputReg := regexp.MustCompile(`<div[^>]*class="output"[^>]*>[\s\S]*?<pre[^>]*>([\s\S]*?)</pre>`)
	outputMatches := outputReg.FindAllSubmatch(body, -1)

	if len(inputMatches) == 0 || len(outputMatches) == 0 {
		return nil, nil, fmt.Errorf("Cannot parse sample with input %v and output %v", len(inputMatches), len(outputMatches))
	}

	count := len(inputMatches)
	if len(outputMatches) < count {
		count = len(outputMatches)
	}

	for i := 0; i < count; i++ {
		inputContent := extractTextContent(inputMatches[i][1])
		input = append(input, []byte(inputContent+"\n"))

		outputContent := extractTextContent(outputMatches[i][1])
		output = append(output, []byte(outputContent+"\n"))
	}

	return input, output, nil
}

// extractTextContent extracts text content from HTML, removing all tags
// Preserves newlines while normalizing spaces and tabs
// Handles both old format (<br> tags) and new format (<div> tags)
func extractTextContent(htmlBytes []byte) string {
	text := string(htmlBytes)

	// STEP 1: Handle <div> tags (NEW format - recent contests)
	// Replace closing </div> tags with newlines to preserve line breaks
	// Each <div>...</div> represents one line in the new format
	divReg := regexp.MustCompile(`</div>`)
	text = divReg.ReplaceAllString(text, "\n")
	logger.Info("[HTML Parser] Replaced </div> tags with newlines (new format)")

	// STEP 2: Handle <br> tags (OLD format - older contests)
	// Replace <br>, <br/>, and <br /> with newlines
	brReg := regexp.MustCompile(`<br\s*/?>`)
	text = brReg.ReplaceAllString(text, "\n")
	logger.Info("[HTML Parser] Replaced <br> tags with newlines (old format)")

	// STEP 3: Remove all remaining HTML tags
	// At this point, all structural tags are gone or replaced with newlines
	tagReg := regexp.MustCompile(`<[^>]+>`)
	text = tagReg.ReplaceAllString(text, "")
	logger.Debug("[HTML Parser] Removed remaining HTML tags")

	// STEP 4: Unescape HTML entities
	text = html.UnescapeString(text)
	logger.Debug("[HTML Parser] Unescaped HTML entities")

	// STEP 5: Normalize horizontal whitespace (spaces and tabs only)
	// Do NOT touch newlines or carriage returns
	spaceReg := regexp.MustCompile(`[ \t]+`)
	text = spaceReg.ReplaceAllString(text, " ")
	logger.Debug("[HTML Parser] Normalized horizontal whitespace")

	// STEP 6: Trim each line and preserve line breaks
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	text = strings.Join(lines, "\n")
	logger.Debug("[HTML Parser] Trimmed each line")

	// STEP 7: Trim leading/trailing whitespace but keep structure
	text = strings.Trim(text, " \t\r")

	logger.Info("[HTML Parser] Extraction complete: %d lines", strings.Count(text, "\n")+1)
	return text
}

// IsStandardIO checks if problem uses standard input/output
func IsStandardIO(body []byte) bool {
	standardIOMarker := []byte(`<div class="input-file"><div class="property-title">input</div>standard input</div><div class="output-file"><div class="property-title">output</div>standard output</div>`)
	return !strings.Contains(string(body), string(standardIOMarker))
}
