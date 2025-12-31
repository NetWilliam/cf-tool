package cmd

import (
	"time"

	"github.com/NetWilliam/cf-tool/client"
	"github.com/NetWilliam/cf-tool/config"
)

// Race command
func Race() (err error) {
	cfg := config.Instance
	cln := client.Instance
	info := Args.Info

	err = cln.RaceContest(info)
	if err != nil {
		return
	}
	time.Sleep(1)
	URL, err := info.ProblemSetURL(cfg.Host)
	if err != nil {
		return
	}
	openURL(URL)
	openURL(URL + "/problems")
	return Parse()
}
