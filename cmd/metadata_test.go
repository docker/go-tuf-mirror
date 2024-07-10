package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	DelegatedTargetNames = [1]string{"test-role"}
)

func TestMetadataCmd(t *testing.T) {
	tempDir := OCIPrefix + os.TempDir()

	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join("..", "internal", "test", "testdata", "test-repo"))))
	defer server.Close()

	testCases := []struct {
		name        string
		source      string
		destination string
		full        bool
	}{
		{"http metadata to oci", server.URL + "/metadata", tempDir, false},
		{"http metadata with delegates to oci", server.URL + "/metadata", tempDir, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedOutput := fmt.Sprintf("Mirroring TUF metadata %s to %s\nMetadata manifest layout saved to %s\n",
				tc.source,
				tc.destination,
				strings.TrimPrefix(tc.destination, OCIPrefix))
			if tc.full {
				for _, d := range DelegatedTargetNames {
					expectedOutput += fmt.Sprintf("Delegated metadata manifest layout saved to %s\n", filepath.Join(strings.TrimPrefix(tc.destination, OCIPrefix), d))
				}
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

			os.RemoveAll(tc.destination)

			out, err := io.ReadAll(b)
			require.NoError(t, err)

			assert.Equal(t, expectedOutput, string(out))
		})
	}
}
