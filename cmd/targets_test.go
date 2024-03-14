package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-tuf-mirror/pkg/mirror"
	"github.com/docker/go-tuf-mirror/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargetsCmd(t *testing.T) {
	tempDir := types.OCIPrefix + os.TempDir()

	testCases := []struct {
		name        string
		source      string
		destination string
		metadata    string
		full        bool
	}{
		{"http targets to oci", mirror.DefaultTargetsURL, tempDir, mirror.DefaultMetadataURL, false},
		{"http targets with delegates to oci", mirror.DefaultTargetsURL, tempDir, mirror.DefaultMetadataURL, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedOutput := fmt.Sprintf("Mirroring TUF targets %s to %s\n", tc.source, tc.destination)

			opts := defaultRootOptions()
			opts.full = tc.full
			cmd := newTargetsCmd(opts)
			if cmd == nil {
				t.Fatal("newTargetsCmd returned nil")
			}
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			_ = cmd.PersistentFlags().Set("source", tc.source)
			_ = cmd.PersistentFlags().Set("destination", tc.destination)
			_ = cmd.PersistentFlags().Set("metadata", tc.metadata)

			err := cmd.Execute()
			require.NoError(t, err)

			os.RemoveAll(tc.destination)

			reader := bufio.NewReader(b)
			out, err := reader.ReadString('\n')
			require.NoError(t, err)

			assert.Equal(t, expectedOutput, out)
		})
	}
}
