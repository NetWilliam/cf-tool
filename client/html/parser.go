package html

import (
	"fmt"
	"html"
	"regexp"
	"strings"
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
func extractTextContent(htmlBytes []byte) string {
	text := string(htmlBytes)

	// CRITICAL: Replace <br> tags with newlines BEFORE removing other tags
	// This handles both <br>, <br/>, and <br /> variants
	brReg := regexp.MustCompile(`<br\s*/?>`)
	text = brReg.ReplaceAllString(text, "\n")

	// Remove all remaining HTML tags
	tagReg := regexp.MustCompile(`<[^>]+>`)
	text = tagReg.ReplaceAllString(text, "")

	// Unescape HTML entities
	text = html.UnescapeString(text)

	// ONLY replace spaces and tabs, NOT newlines or carriage returns
	// This preserves line breaks while normalizing horizontal whitespace
	spaceReg := regexp.MustCompile(`[ \t]+`)
	text = spaceReg.ReplaceAllString(text, " ")

	// Trim each line and preserve line breaks
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	text = strings.Join(lines, "\n")

	// Trim leading/trailing whitespace (spaces/tabs) but keep structure
	text = strings.Trim(text, " \t\r")

	return text
}

// IsStandardIO checks if problem uses standard input/output
func IsStandardIO(body []byte) bool {
	standardIOMarker := []byte(`<div class="input-file"><div class="property-title">input</div>standard input</div><div class="output-file"><div class="property-title">output</div>standard output</div>`)
	return !strings.Contains(string(body), string(standardIOMarker))
}
