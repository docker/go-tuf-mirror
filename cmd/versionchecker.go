package cmd

import (
	"github.com/docker/attest/tuf"
)

type nullVersionChecker struct{}

func (*nullVersionChecker) CheckVersion(tuf.Downloader) error {
	return nil
}
