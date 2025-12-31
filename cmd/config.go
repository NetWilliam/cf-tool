package cmd

import (
	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/NetWilliam/cf-tool/config"
	"github.com/NetWilliam/cf-tool/util"
)

// Config command
func Config() (err error) {
	cfg := config.Instance
	color.Cyan("Configure the tool")
	ansi.Println(`0) add a template`)
	ansi.Println(`1) delete a template`)
	ansi.Println(`2) set default template`)
	ansi.Println(`3) run "cf gen" after "cf parse"`)
	ansi.Println(`4) set host domain`)
	ansi.Println(`5) set proxy`)
	ansi.Println(`6) set folders' name`)
	index := util.ChooseIndex(7)
	if index == 0 {
		return cfg.AddTemplate()
	} else if index == 1 {
		return cfg.RemoveTemplate()
	} else if index == 2 {
		return cfg.SetDefaultTemplate()
	} else if index == 3 {
		return cfg.SetGenAfterParse()
	} else if index == 4 {
		return cfg.SetHost()
	} else if index == 5 {
		return cfg.SetProxy()
	} else if index == 6 {
		return cfg.SetFolderName()
	}
	return
}
