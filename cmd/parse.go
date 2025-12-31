package cmd

import (
	"errors"
	"path/filepath"

	"github.com/NetWilliam/cf-tool/client"
	"github.com/NetWilliam/cf-tool/config"
)

// Parse command
func Parse() (err error) {
	cfg := config.Instance
	cln := client.Instance
	info := Args.Info
	source := ""
	ext := ""
	if cfg.GenAfterParse {
		if len(cfg.Template) == 0 {
			return errors.New("You have to add at least one code template by `cf config`")
		}
		path := cfg.Template[cfg.Default].Path
		ext = filepath.Ext(path)
		if source, err = readTemplateSource(path, cln); err != nil {
			return
		}
	}
	work := func() error {
		_, paths, err := cln.Parse(info)
		if err != nil {
			return err
		}
		if cfg.GenAfterParse {
			for _, path := range paths {
				gen(source, path, ext)
			}
		}
		return nil
	}
	return executeWithLoginRetry(cln, work)
}
