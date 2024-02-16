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

func TestAll(t *testing.T) {
	tempDir := os.TempDir()
	tempPath := types.OCIPrefix + os.TempDir()
	testCases := []struct {
		name    string
		srcMeta string
		dstMeta string
		srcTgt  string
		dstTgt  string
	}{
		{"git targets to oci", mirror.DefaultMetadataURL, tempPath, mirror.DefaultTargetsURL, tempPath},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := defaultRootOptions()
			opts.tufPath = tempDir
			cmd := newAllCmd(opts)

			expectedMetadataOutput := fmt.Sprintf("Mirroring TUF metadata %s to %s\n", tc.srcMeta, tc.dstMeta)
			expectedTargetsOutput := fmt.Sprintf("Mirroring TUF targets %s to %s\n", tc.srcTgt, tc.dstTgt)

			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			_ = cmd.Flags().Set("source-metadata", tc.srcMeta)
			_ = cmd.Flags().Set("source-targets", tc.srcTgt)
			_ = cmd.Flags().Set("dest-metadata", tc.dstMeta)
			_ = cmd.Flags().Set("dest-targets", tc.dstTgt)

			err := cmd.Execute()
			require.NoError(t, err)

			os.RemoveAll("./tmp")

			reader := bufio.NewReader(b)
			metaOut, err := reader.ReadString('\n')
			require.NoError(t, err)
			assert.Equal(t, expectedMetadataOutput, metaOut)

			_, err = reader.ReadString('\n')
			require.NoError(t, err)
			targetsOut, err := reader.ReadString('\n')
			require.NoError(t, err)
			assert.Equal(t, expectedTargetsOutput, targetsOut)
		})
	}
}
