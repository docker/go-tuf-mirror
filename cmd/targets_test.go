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
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	targetFile = "02119a076ec3878c736c3a95e20794f5a8d5bce3d7ecc264681bb7334ca2e24b.test.txt"
)

func TestTargetsCmd(t *testing.T) {
	tempDir := OCIPrefix + filepath.Join(os.TempDir(), "test")

	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join("..", "internal", "test", "testdata", "test-repo"))))
	defer server.Close()
	serverMetadata := server.URL + "/metadata"
	serverTargets := server.URL + "/targets"

	reg := httptest.NewServer(registry.New(registry.WithReferrersSupport(false)))
	defer reg.Close()
	url, err := url.Parse(reg.URL)
	require.NoError(t, err)
	registryPath := RegistryPrefix + "localhost:" + url.Port() + "/test/targets"

	testCases := []struct {
		name        string
		source      string
		destination string
		metadata    string
		full        bool
	}{
		{"http targets to oci", serverTargets, tempDir, serverMetadata, false},
		{"http targets with delegates to oci", serverTargets, tempDir, serverMetadata, true},
		{"http metadata to registry", serverTargets, registryPath, serverMetadata, false},
		{"http metadata with delegates to registry", serverTargets, registryPath, serverMetadata, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedOutput := fmt.Sprintf("Mirroring TUF targets %s to %s\n", tc.source, tc.destination)

			opts := defaultRootOptions()
			opts.full = tc.full
			opts.tufRoot = "dev"
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

			reader := bufio.NewReader(b)
			out, err := reader.ReadString('\n')
			require.NoError(t, err)
			assert.Equal(t, expectedOutput, out)

			// check that index was saved to oci layout
			if strings.HasPrefix(tc.destination, OCIPrefix) {
				data, err := os.ReadFile(filepath.Join(strings.TrimPrefix(tc.destination, OCIPrefix), filepath.Join(targetFile, "index.json")))
				require.NoError(t, err)
				assert.True(t, len(data) > 0)
				err = os.RemoveAll(strings.TrimPrefix(tc.destination, OCIPrefix))
				require.NoError(t, err)
			}

			// check that image was pushed to registry
			if strings.HasPrefix(tc.destination, RegistryPrefix) {
				ref, err := name.ParseReference(strings.TrimPrefix(strings.Join([]string{tc.destination, targetFile}, ":"), RegistryPrefix))
				require.NoError(t, err)
				image, err := remote.Image(ref)
				require.NoError(t, err)
				size, err := image.Size()
				require.NoError(t, err)
				assert.True(t, size > 0)
			}
		})
	}
}
