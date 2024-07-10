package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	DelegatedTargetsLength = 1
)

func TestAll(t *testing.T) {
	tempDir := os.TempDir()
	tempPath := OCIPrefix + os.TempDir()

	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join("..", "internal", "test", "testdata", "test-repo"))))
	defer server.Close()

	testCases := []struct {
		name    string
		srcMeta string
		dstMeta string
		srcTgt  string
		dstTgt  string
		full    bool
	}{
		{"http to oci", server.URL + "/metadata", tempPath, server.URL + "/targets", tempPath, false},
		{"http with delegates to oci", server.URL + "/metadata", tempPath, server.URL + "/targets", tempPath, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := defaultRootOptions()
			opts.tufPath = tempDir
			opts.full = tc.full
			opts.tufRoot = "dev"
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
			if tc.full {
				for i := 0; i < DelegatedTargetsLength; i++ {
					_, err = reader.ReadString('\n')
					require.NoError(t, err)
				}
			}
			targetsOut, err := reader.ReadString('\n')
			require.NoError(t, err)
			assert.Equal(t, expectedTargetsOutput, targetsOut)
		})
	}
}
