package tuf

import (
	_ "embed"
	"os"
	"testing"
)

//go:embed 1.root-staging.json
var InitialRoot []byte

// NewTufClient creates a new TUF client
func TestRootInit(t *testing.T) {
	tufPath, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewTufClient(InitialRoot, tufPath, "https://docker.github.io/tuf-staging/metadata", "https://docker.github.io/tuf-staging/targets")
	if err != nil {
		t.Fatal("Failed to create TUF client: ", err)
	}
	// recreation should work with same root
	_, err = NewTufClient(InitialRoot, tufPath, "https://docker.github.io/tuf-staging/metadata", "https://docker.github.io/tuf-staging/targets")
	if err != nil {
		t.Fatal("Failed to recreate TUF client:", err)
	}
	_, err = NewTufClient([]byte("broken"), tufPath, "https://docker.github.io/tuf-staging/metadata", "https://docker.github.io/tuf-staging/targets")
	if err == nil {
		t.Fatal("Expected error recreating TUF client with broken root")
	}
}
