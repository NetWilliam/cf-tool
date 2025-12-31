package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/docopt/docopt-go"

	"github.com/NetWilliam/cf-tool/client"
	"github.com/NetWilliam/cf-tool/config"
	"github.com/NetWilliam/cf-tool/pkg/mcp"
	"github.com/fatih/color"
	"github.com/NetWilliam/cf-tool/util"
)

// Eval opts
func Eval(opts docopt.Opts) error {
	Args = &ParsedArgs{}
	opts.Bind(Args)
	if err := parseArgs(opts); err != nil {
		return err
	}
	if Args.Config {
		return Config()
	} else if Args.Submit {
		return Submit()
	} else if Args.List {
		return List()
	} else if Args.Parse {
		return Parse()
	} else if Args.Gen {
		return Gen()
	} else if Args.Test {
		return Test()
	} else if Args.Watch {
		return Watch()
	} else if Args.Open {
		return Open()
	} else if Args.Stand {
		return Stand()
	} else if Args.Sid {
		return Sid()
	} else if Args.Race {
		return Race()
	} else if Args.Pull {
		return Pull()
	} else if Args.Clone {
		return Clone()
	} else if Args.Upgrade {
		return Upgrade()
	} else if Args.McpPing {
		return McpPing()
	} else if Args.Mocka {
		return Mocka()
	} else if Args.LogTest {
		return LogTest()
	}
	return nil
}

func getSampleID() (samples []string) {
	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := os.ReadDir(path)
	if err != nil {
		return
	}
	reg := regexp.MustCompile(`in(\d+).txt`)
	for _, path := range paths {
		name := path.Name()
		tmp := reg.FindSubmatch([]byte(name))
		if tmp != nil {
			idx := string(tmp[1])
			ans := fmt.Sprintf("ans%v.txt", idx)
			if _, err := os.Stat(ans); err == nil {
				samples = append(samples, idx)
			}
		}
	}
	return
}

// CodeList Name matches some template suffix, index are template array indexes
type CodeList struct {
	Name  string
	Index []int
}

func getCode(filename string, templates []config.CodeTemplate) (codes []CodeList, err error) {
	mp := make(map[string][]int)
	for i, temp := range templates {
		suffixMap := map[string]bool{}
		for _, suffix := range temp.Suffix {
			if _, ok := suffixMap[suffix]; !ok {
				suffixMap[suffix] = true
				sf := "." + suffix
				mp[sf] = append(mp[sf], i)
			}
		}
	}

	if filename != "" {
		ext := filepath.Ext(filename)
		if idx, ok := mp[ext]; ok {
			return []CodeList{CodeList{filename, idx}}, nil
		}
		return nil, fmt.Errorf("%v can not match any template. You could add a new template by `cf config`", filename)
	}

	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := os.ReadDir(path)
	if err != nil {
		return
	}

	for _, path := range paths {
		name := path.Name()
		ext := filepath.Ext(name)
		if idx, ok := mp[ext]; ok {
			codes = append(codes, CodeList{name, idx})
		}
	}

	return codes, nil
}

func getOneCode(filename string, templates []config.CodeTemplate) (name string, index int, err error) {
	codes, err := getCode(filename, templates)
	if err != nil {
		return
	}
	if len(codes) < 1 {
		return "", 0, errors.New("Cannot find any code.\nMaybe you should add a new template by `cf config`")
	}
	if len(codes) > 1 {
		color.Cyan("There are multiple files can be selected.")
		for i, code := range codes {
			fmt.Printf("%3v: %v\n", i, code.Name)
		}
		i := util.ChooseIndex(len(codes))
		codes[0] = codes[i]
	}
	if len(codes[0].Index) > 1 {
		color.Cyan("There are multiple languages match the file.")
		for i, idx := range codes[0].Index {
			fmt.Printf("%3v: %v\n", i, client.Langs[templates[idx].Lang])
		}
		i := util.ChooseIndex(len(codes[0].Index))
		codes[0].Index[0] = codes[0].Index[i]
	}
	return codes[0].Name, codes[0].Index[0], nil
}

func loginAgain(cln *client.Client, err error) error {
	if err != nil && err.Error() == client.ErrorNotLogged {
		color.Red("Not logged. Try to login\n")

		// Check if password is configured
		if len(cln.Password) == 0 || len(cln.HandleOrEmail) == 0 {
			// No password configured, try browser mode
			color.Yellow("No password configured. Attempting browser mode login...\n")

			// Try to initialize browser client
			if browserErr := tryInitBrowserClient(cln); browserErr == nil {
				// Browser mode initialized successfully, try login
				return cln.Login()
			} else {
				// Browser mode failed, prompt user to configure
				color.Red("\n❌ Browser mode is not available.\n")
				color.Yellow("Please configure your handle and password:\n")
				color.Cyan("  Option 1: cf config\n")
				color.Cyan("  Option 2: Install MCP Chrome Server for browser mode\n")
				return errors.New("Not logged in. Please configure credentials or enable browser mode")
			}
		}

		// Password is configured, proceed with normal login
		err = cln.Login()
	}
	return err
}

// tryInitBrowserClient attempts to initialize browser client
func tryInitBrowserClient(cln *client.Client) error {
	// Try to find MCP server
	serverURL, mcpPath, err := findMCPServer()
	if err != nil {
		return fmt.Errorf("MCP server not found: %w", err)
	}

	var mcpClient *mcp.Client

	// Determine which transport to use
	if serverURL != "" {
		color.Cyan("Using HTTP transport: %s\n", serverURL)
		mcpClient, err = mcp.NewClientHTTP(serverURL)
	} else if mcpPath != "" {
		color.Cyan("Using stdio transport: %s\n", mcpPath)
		mcpClient, err = mcp.NewClient("node", []string{mcpPath})
	} else {
		return fmt.Errorf("no valid MCP server configuration found")
	}

	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}

	// Set MCP client and enable browser mode
	cln.SetMCPClient(mcpClient)

	return nil
}
