package mirror

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/docker/go-tuf-mirror/internal/tuf"
	"github.com/theupdateframework/go-tuf/v2/metadata"
)

//go:embed 1.root-staging.json
var InitialRoot []byte

const (
	defaultMetadataURL = "https://docker.github.io/tuf-staging/metadata"
	defaultTargetsURL  = "https://docker.github.io/tuf-staging/targets"
)

type TufMetadata struct {
	Root      []byte
	Snapshot  []byte
	Targets   []byte
	Timestamp []byte
}

func GetTufGitRepoMetadata() (*TufMetadata, error) {
	tufPath, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	tufClient, err := tuf.NewTufClient(InitialRoot, tufPath, defaultMetadataURL, defaultTargetsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUF client: %w", err)
	}

	trustedMetadata := tufClient.GetMetadata()

	rootBytes, err := trustedMetadata.Root.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get root metadata: %w", err)
	}
	snapshotBytes, err := trustedMetadata.Snapshot.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot metadata: %w", err)
	}
	targetsBytes, err := trustedMetadata.Targets[metadata.TARGETS].ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get targets metadata: %w", err)
	}
	timstampBytes, err := trustedMetadata.Timestamp.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get timestamp metadata: %w", err)
	}
	return &TufMetadata{
		Root:      rootBytes,
		Snapshot:  snapshotBytes,
		Targets:   targetsBytes,
		Timestamp: timstampBytes,
	}, nil
}

func PushMetadataManifest() error {
	metadata, err := GetTufGitRepoMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}
	fmt.Printf("metadata: %s\n", metadata)
	return nil
}
