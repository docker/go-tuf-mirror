/*
   Copyright Docker go-tuf-mirror authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cmd

import (
	"bytes"
	"fmt"
	"io"
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

var (
	DelegatedTargetNames = [1]string{"test-role"}
)

func TestMetadataCmd(t *testing.T) {
	tempDir := OCIPrefix + filepath.Join(os.TempDir(), "test")

	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join("..", "internal", "test", "testdata", "test-repo"))))
	defer server.Close()
	serverMetadata := server.URL + "/metadata"

	reg := httptest.NewServer(registry.New(registry.WithReferrersSupport(false)))
	defer reg.Close()
	url, err := url.Parse(reg.URL)
	require.NoError(t, err)
	registryPath := RegistryPrefix + "localhost:" + url.Port() + "/test/metadata:latest"

	testCases := []struct {
		name        string
		source      string
		destination string
		full        bool
	}{
		{"http metadata to oci", serverMetadata, tempDir, false},
		{"http metadata with delegates to oci", serverMetadata, tempDir, true},
		{"http metadata to registry", serverMetadata, registryPath, false},
		{"http metadata with delegates to registry", serverMetadata, registryPath, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var operation string
			var output string
			var delegatedOutput string
			if strings.HasPrefix(tc.destination, RegistryPrefix) {
				operation = "pushed"
				ref, err := name.ParseReference(strings.TrimPrefix(tc.destination, RegistryPrefix))
				require.NoError(t, err)
				output = ref.Name()
				for _, d := range DelegatedTargetNames {
					delegatedOutput += fmt.Sprintf("Delegated metadata manifest %s to %s\n", operation, strings.Join([]string{ref.Context().Name(), d}, ":"))
				}
			} else {
				operation = "layout saved"
				output = strings.TrimPrefix(tc.destination, OCIPrefix)
				for _, d := range DelegatedTargetNames {
					delegatedOutput += fmt.Sprintf("Delegated metadata manifest %s to %s\n", operation, filepath.Join(output, d))
				}
			}
			expectedOutput := fmt.Sprintf("Mirroring TUF metadata %s to %s\nMetadata manifest %s to %s\n",
				tc.source,
				tc.destination,
				operation,
				output)
			if tc.full {
				expectedOutput += delegatedOutput
			}

			b := bytes.NewBufferString("")
			opts := defaultRootOptions()
			opts.full = tc.full
			opts.tufRoot = "dev"
			cmd := newMetadataCmd(opts)
			if cmd == nil {
				t.Fatal("newMetadataCmd returned nil")
			}
			cmd.SetOut(b)
			_ = cmd.PersistentFlags().Set("source", tc.source)
			_ = cmd.PersistentFlags().Set("destination", tc.destination)

			err := cmd.Execute()
			require.NoError(t, err)

			out, err := io.ReadAll(b)
			require.NoError(t, err)
			assert.Equal(t, expectedOutput, string(out))

			// check that index was saved to oci layout
			if strings.HasPrefix(tc.destination, OCIPrefix) {
				data, err := os.ReadFile(filepath.Join(strings.TrimPrefix(tc.destination, OCIPrefix), "index.json"))
				require.NoError(t, err)
				assert.True(t, len(data) > 0)
				err = os.RemoveAll(strings.TrimPrefix(tc.destination, OCIPrefix))
				require.NoError(t, err)
			}

			// check that image was pushed to registry
			if strings.HasPrefix(tc.destination, RegistryPrefix) {
				ref, err := name.ParseReference(strings.TrimPrefix(tc.destination, RegistryPrefix))
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
