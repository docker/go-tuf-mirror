package mirror

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/docker/go-tuf-mirror/internal/tuf"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/theupdateframework/go-tuf/v2/metadata"
	"github.com/theupdateframework/go-tuf/v2/metadata/fetcher"
)

//go:embed 1.root-staging.json
var InitialRoot []byte

const (
	defaultMetadataURL   = "https://docker.github.io/tuf-staging/metadata"
	defaultTargetsURL    = "https://docker.github.io/tuf-staging/targets"
	tufMetadataMediaType = "application/vnd.tuf.metadata+json"
	tufRoleAnnotation    = "tuf.io/role"
)

type TufRole string

var TufRoles = []TufRole{metadata.ROOT, metadata.SNAPSHOT, metadata.TARGETS, metadata.TIMESTAMP}

type TufMetadata struct {
	Root      map[string][]byte
	Snapshot  map[string][]byte
	Targets   map[string][]byte
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
	// if trustedMetadata.Root.Signed.ConsistentSnapshot {
	// 	// TODO - implement consistent snapshot metadata
	// 	return nil, fmt.Errorf("consistent snapshot metadata not implemented")
	// }

	rootMetadata := map[string][]byte{}
	rootBytes, err := trustedMetadata.Root.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get root metadata: %w", err)
	}
	rootVersion := trustedMetadata.Root.Signed.Version
	rootMetadata[fmt.Sprintf("%d.root.json", rootVersion)] = rootBytes
	// get the previous versions of root metadata if any
	if trustedMetadata.Root.Signed.Version != 1 {
		client := fetcher.DefaultFetcher{}
		for i := 1; i < int(trustedMetadata.Root.Signed.Version); i++ {
			meta, err := client.DownloadFile(defaultMetadataURL+fmt.Sprintf("/%d.root.json", i), tufClient.MaxRootLength(), time.Second*15)
			if err != nil {
				return nil, fmt.Errorf("failed to download root metadata: %w", err)
			}
			rootMetadata[fmt.Sprintf("%d.root.json", i)] = meta
		}
	}

	snapshotBytes, err := trustedMetadata.Snapshot.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot metadata: %w", err)
	}
	targetsBytes, err := trustedMetadata.Targets[metadata.TARGETS].ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get targets metadata: %w", err)
	}
	timestampBytes, err := trustedMetadata.Timestamp.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get timestamp metadata: %w", err)
	}
	return &TufMetadata{
		Root:      rootMetadata,
		Snapshot:  map[string][]byte{"snapshot.json": snapshotBytes},
		Targets:   map[string][]byte{"targets.json": targetsBytes},
		Timestamp: timestampBytes,
	}, nil
}

func BuildMetadataManifest(metadata *TufMetadata) (v1.Image, error) {
	img := empty.Image
	img = mutate.MediaType(img, types.OCIManifestSchema1)
	img = mutate.ConfigMediaType(img, types.OCIConfigJSON)
	for _, role := range TufRoles {
		layers, err := makeRoleLayer(role, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to make role layer: %w", err)
		}
		img, err = mutate.Append(img, *layers...)
		if err != nil {
			return nil, fmt.Errorf("failed to append role layer to image: %w", err)
		}
	}
	return img, nil
}

func makeRoleLayer(role TufRole, tufMetadata *TufMetadata) (*[]mutate.Addendum, error) {
	layers := new([]mutate.Addendum)
	switch role {
	case metadata.ROOT:
		for name, data := range tufMetadata.Root {
			*layers = append(*layers, mutate.Addendum{Layer: static.NewLayer(data, tufMetadataMediaType), Annotations: map[string]string{tufRoleAnnotation: name}})
		}
	case metadata.SNAPSHOT:
		for name, data := range tufMetadata.Snapshot {
			*layers = append(*layers, mutate.Addendum{Layer: static.NewLayer(data, tufMetadataMediaType), Annotations: map[string]string{tufRoleAnnotation: name}})
		}
	case metadata.TARGETS:
		for name, data := range tufMetadata.Targets {
			*layers = append(*layers, mutate.Addendum{Layer: static.NewLayer(data, tufMetadataMediaType), Annotations: map[string]string{tufRoleAnnotation: name}})
		}
	case metadata.TIMESTAMP:
		*layers = append(*layers, mutate.Addendum{Layer: static.NewLayer(tufMetadata.Timestamp, tufMetadataMediaType), Annotations: map[string]string{tufRoleAnnotation: string(role)}})
	default:
		return nil, fmt.Errorf("unsupported TUF role: %s", role)
	}
	return layers, nil
}

func PushMetadataManifest(imageName string) error {
	metadata, err := GetTufGitRepoMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}
	manifest, err := BuildMetadataManifest(metadata)
	if err != nil {
		return fmt.Errorf("failed to build metadata manifest: %w", err)
	}
	err = PushToRegistry(manifest, imageName)
	if err != nil {
		return fmt.Errorf("failed to push metadata manifest: %w", err)
	}
	return nil
}

func PushToRegistry(image v1.Image, imageName string) error {
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
	if err := remote.Write(ref, image, remote.WithAuth(auth)); err != nil {
		return fmt.Errorf("failed to push image %s: %w", imageName, err)
	}
	return nil
}
