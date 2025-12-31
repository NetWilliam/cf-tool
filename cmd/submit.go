package cmd

import (
	"os"

	"github.com/NetWilliam/cf-tool/client"
	"github.com/NetWilliam/cf-tool/config"
)

// Submit command
func Submit() (err error) {
	cln := client.Instance
	cfg := config.Instance
	info := Args.Info
	filename, index, err := getOneCode(Args.File, cfg.Template)
	if err != nil {
		return
	}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	source := string(bytes)

	lang := cfg.Template[index].Lang
	return cln.Submit(info, lang, source)
}
