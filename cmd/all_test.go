package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	DelegatedTargetsLength = 1
)

func TestAll(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test")
	tempPath := OCIPrefix + tempDir

	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join("..", "internal", "test", "testdata", "test-repo"))))
	defer server.Close()
	serverMetadata := server.URL + "/metadata"
	serverTargets := server.URL + "/targets"

	reg := httptest.NewServer(registry.New(registry.WithReferrersSupport(false)))
	defer reg.Close()
	url, err := url.Parse(reg.URL)
	require.NoError(t, err)
	registryPathMetadata := RegistryPrefix + "localhost:" + url.Port() + "/test/metadata:latest"
	registryPathTargets := RegistryPrefix + "localhost:" + url.Port() + "/test/targets"

	testCases := []struct {
		name    string
		srcMeta string
		dstMeta string
		srcTgt  string
		dstTgt  string
		full    bool
	}{
		{"http to oci", serverMetadata, tempPath, serverTargets, tempPath, false},
		{"http with delegates to oci", serverMetadata, tempPath, serverTargets, tempPath, true},
		{"http metadata to registry", serverMetadata, registryPathMetadata, serverTargets, registryPathTargets, false},
		{"http metadata with delegates to registry", serverMetadata, registryPathMetadata, serverTargets, registryPathTargets, true},
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

			err = os.RemoveAll("./tmp")
			require.NoError(t, err)

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
