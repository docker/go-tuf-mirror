package tuf

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/test-repo/metadata/1.root.json
var TestRoot []byte

// NewTufClient creates a new TUF client
func TestRootInit(t *testing.T) {
	tufPath, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	// Start a test HTTP server to serve data from ./testdata/test-repo/ paths
	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join(".", "testdata", "test-repo"))))
	defer server.Close()

	testCases := []struct {
		name           string
		metadataSource string
		targetsSource  string
	}{
		{"http", server.URL + "/metadata", server.URL + "/targets"},
	}

	for _, tc := range testCases {
		_, err = NewTufClient(TestRoot, tufPath, tc.metadataSource, tc.targetsSource)
		if err != nil {
			t.Fatal("Failed to create TUF client: ", err)
		}
		// recreation should work with same root
		_, err = NewTufClient(TestRoot, tufPath, tc.metadataSource, tc.targetsSource)
		if err != nil {
			t.Fatal("Failed to recreate TUF client:", err)
		}
		_, err = NewTufClient([]byte("broken"), tufPath, tc.metadataSource, tc.targetsSource)
		if err == nil {
			t.Fatal("Expected error recreating TUF client with broken root")
		}
	}
}
