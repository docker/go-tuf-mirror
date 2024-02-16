package test

import (
	"os"
	"testing"
)

func CreateTempDir(t *testing.T, pattern string) string {
	// Create a temporary directory for output oci layout
	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Register a cleanup function to delete the temp directory when the test exits
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	})
	return tempDir
}
