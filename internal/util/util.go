package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

func HexHashBytes(input []byte) string {
	s256 := sha256.New()
	s256.Write(input)
	hashSum := s256.Sum(nil)
	return hex.EncodeToString(hashSum)
}

// isValidUrl tests a string to determine if it is a well-structured url or not.
func IsValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func CreateEmptyIndex() (v1.ImageIndex, error) {
	// Create a temporary directory for output oci layout
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	tempDir, err := os.MkdirTemp(dir, "isv-wrapper")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	defer os.RemoveAll(tempDir) // clean up
	tempIndex := empty.Index
	_, err = layout.Write(tempDir, tempIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to write signed image: %w", err)
	}
	return layout.ImageIndexFromPath(tempDir)
}
