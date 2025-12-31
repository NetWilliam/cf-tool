package cmd

import (
	"os"

	"github.com/NetWilliam/cf-tool/client"
)

// Pull command
func Pull() (err error) {
	cln := client.Instance
	info := Args.Info
	ac := Args.Accepted
	rootPath, err := os.Getwd()
	if err != nil {
		return
	}
	return executeWithLoginRetry(cln, func() error {
		return cln.Pull(info, rootPath, ac)
	})
}
