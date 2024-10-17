package cmd

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	av "github.com/docker/attest/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	version := "v1.0.0"
	cmd := newVersionCmd(version)
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	err := cmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(b)
	require.NoError(t, err)
	fetcher := av.NewGoVersionFetcher()
	attestVersion, err := fetcher.Get()
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("github.com/docker/go-tuf-mirror: %s\ngithub.com/docker/attest: %s\n", version, attestVersion), string(out))
}
