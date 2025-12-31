package cmd

import (
	"os"

	"github.com/NetWilliam/cf-tool/client"
)

// Clone command
func Clone() (err error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return
	}
	cln := client.Instance
	ac := Args.Accepted
	handle := Args.Handle

	err = cln.Clone(handle, currentPath, ac)
	return
}
