package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/docker/go-tuf-mirror/pkg/mirror"
	"github.com/docker/go-tuf-mirror/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataCmd(t *testing.T) {
	opts := defaultRootOptions()
	cmd := newMetadataCmd(opts)
	if cmd == nil {
		t.Fatal("newMetadataCmd returned nil")
	}

	tempDir := types.OCIPrefix + os.TempDir()

	testCases := []struct {
		name        string
		source      string
		destination string
	}{
		{"git metadata to oci", mirror.DefaultMetadataURL, tempDir},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedOutput := fmt.Sprintf("Mirroring TUF metadata %s to %s\nMetadata manifest layout saved to %s\n",
				tc.source,
				tc.destination,
				strings.TrimPrefix(tc.destination, types.OCIPrefix))

			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			_ = cmd.PersistentFlags().Set("source", tc.source)
			_ = cmd.PersistentFlags().Set("destination", tc.destination)

			err := cmd.Execute()
			require.NoError(t, err)

			os.RemoveAll(tc.destination)

			out, err := io.ReadAll(b)
			require.NoError(t, err)

			assert.Equal(t, expectedOutput, string(out))
		})
	}
}
