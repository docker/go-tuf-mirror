package mirror

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/docker/go-tuf-mirror/internal/tuf"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func NewTufMirror(tufPath string, metadataURL string, targetsURL string) (*TufMirror, error) {
	tufClient, err := tuf.NewTufClient(InitialRoot, tufPath, metadataURL, targetsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUF client: %w", err)
	}
	return &TufMirror{TufClient: tufClient, tufPath: tufPath, metadataURL: metadataURL, targetsURL: targetsURL}, nil
}

func PushToRegistry(image *v1.Image, imageName string) error {
	// Parse the image name
	ref, err := name.ParseReference(imageName)
	if err != nil {
		log.Fatalf("Failed to parse image name: %v", err)
	}
	// Get the authenticator from the default Docker keychain
	auth, err := authn.DefaultKeychain.Resolve(ref.Context())
	if err != nil {
		log.Fatalf("Failed to get authenticator: %v", err)
	}
	// Push the image to the registry
	if err := remote.Write(ref, *image, remote.WithAuth(auth)); err != nil {
		return fmt.Errorf("failed to push image %s: %w", imageName, err)
	}
	return nil
}

func SaveAsOCILayout(image *v1.Image, path string) error {
	// Save the image to the local filesystem
	err := os.MkdirAll(path, os.FileMode(0744))
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	index := empty.Index
	l, err := layout.Write(path, index)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	err = l.AppendImage(*image)
	if err != nil {
		return fmt.Errorf("failed to append image to index: %w", err)
	}
	return nil
}
