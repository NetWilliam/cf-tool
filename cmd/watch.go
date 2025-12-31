package cmd

import (
	"github.com/NetWilliam/cf-tool/client"
)

// Watch command
func Watch() (err error) {
	cln := client.Instance
	info := Args.Info
	n := 10
	if Args.All {
		n = -1
	}
	return executeWithLoginRetry(cln, func() error {
		_, err := cln.WatchSubmission(info, n, false)
		return err
	})
}
