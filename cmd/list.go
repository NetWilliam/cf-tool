package cmd

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/NetWilliam/cf-tool/client"
	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// List command
func List() (err error) {
	cln := client.Instance
	info := Args.Info

	problems, err := cln.Statis(info)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	output := io.Writer(&buf)

	// Create table with ASCII borders for maximum terminal compatibility
	table := tablewriter.NewTable(output,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleASCII),
		})),
	)

	// Configure table with proper alignment for better column display
	// Column 0 (#): Left-aligned (problem identifier)
	// Column 1 (PROBLEM): Left-aligned (problem name)
	// Column 2 (PASSED): Right-aligned (numeric data)
	// Column 3 (LIMIT): Left-aligned (time/memory limits)
	// Column 4 (IO): Left-aligned (I/O specifications)
	table.Configure(func(config *tablewriter.Config) {
		// Configure header alignment
		config.Header.Alignment.PerColumn = []tw.Align{
			tw.AlignLeft,  // #
			tw.AlignLeft,  // PROBLEM
			tw.AlignRight, // PASSED (numeric column)
			tw.AlignLeft,  // LIMIT
			tw.AlignLeft,  // IO
		}

		// Configure row alignment
		config.Row.Alignment.PerColumn = []tw.Align{
			tw.AlignLeft,  // #
			tw.AlignLeft,  // PROBLEM
			tw.AlignRight, // PASSED (numeric column)
			tw.AlignLeft,  // LIMIT
			tw.AlignLeft,  // IO
		}

		// Set column widths to ensure proper borders and fit in most terminals
		if config.Widths.PerColumn == nil {
			config.Widths.PerColumn = make(tw.Mapper[int, int])
		}
		config.Widths.PerColumn[0] = 3  // # column - fixed width for borders
		config.Widths.PerColumn[1] = 26 // PROBLEM column max width
		config.Widths.PerColumn[2] = 8  // PASSED column width (enough for "PASSED")
		config.Widths.PerColumn[3] = 13 // LIMIT column width
		config.Widths.PerColumn[4] = 18 // IO column width

		// Set maximum table width to fit comfortably in most terminals
		config.MaxWidth = 100
	})

	// Set column headers
	table.Header("#", "PROBLEM", "PASSED", "LIMIT", "IO")

	// Append problem data
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
		if i >= 0 && i < len(problems) {
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
