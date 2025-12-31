package cmd

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/olekukonko/tablewriter"
	"github.com/NetWilliam/cf-tool/client"
)

// List command
func List() (err error) {
	cln := client.Instance
	info := Args.Info

	var problems []client.StatisInfo
	err = executeWithLoginRetry(cln, func() error {
		var e error
		problems, e = cln.Statis(info)
		return e
	})
	if err != nil {
		return
	}

	var buf bytes.Buffer
	output := io.Writer(&buf)
	table := tablewriter.NewWriter(output)
	table.Header("#", "problem", "passed", "limit", "IO")
	for _, prob := range problems {
		table.Append(
			prob.ID,
			prob.Name,
			prob.Passed,
			prob.Limit,
			prob.IO,
		)
	}
	table.Render()

	scanner := bufio.NewScanner(io.Reader(&buf))
	for i := -2; scanner.Scan(); i++ {
		line := scanner.Text()
		if i >= 0 {
			if strings.Contains(problems[i].State, "accepted") {
				line = color.New(color.BgGreen).Sprint(line)
			} else if strings.Contains(problems[i].State, "rejected") {
				line = color.New(color.BgRed).Sprint(line)
			}
		}
		ansi.Println(line)
	}
	return
}
