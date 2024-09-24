package tuf

import (
	"github.com/docker/attest/tuf"
)

type NullVersionChecker struct{}

func (*NullVersionChecker) CheckVersion(tuf.Downloader) error {
	return nil
}
