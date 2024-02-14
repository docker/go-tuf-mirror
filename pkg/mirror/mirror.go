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
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/theupdateframework/go-tuf/v2/metadata"
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

	rootMetadata := map[string][]byte{}
	rootVersion := trustedMetadata.Root.Signed.Version
	// get the previous versions of root metadata if any
	if rootVersion != 1 {
		rootMetadata, err = tufClient.GetPriorRoots(defaultMetadataURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get prior root metadata: %w", err)
		}
	}
	// get current root metadata
	rootBytes, err := trustedMetadata.Root.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get root metadata: %w", err)
	}
	rootMetadata[fmt.Sprintf("%d.root.json", rootVersion)] = rootBytes

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

	snapshotName := "snapshot.json"
	targetsName := "targets.json"
	if trustedMetadata.Root.Signed.ConsistentSnapshot {
		snapshotName = fmt.Sprintf("%d.snapshot.json", trustedMetadata.Snapshot.Signed.Version)
		targetsName = fmt.Sprintf("%d.targets.json", trustedMetadata.Targets[metadata.TARGETS].Signed.Version)
	}
	return &TufMetadata{
		Root:      rootMetadata,
		Snapshot:  map[string][]byte{snapshotName: snapshotBytes},
		Targets:   map[string][]byte{targetsName: targetsBytes},
		Timestamp: timestampBytes,
	}, nil
}

func buildMetadataManifest(metadata *TufMetadata) (v1.Image, error) {
	img := empty.Image
	img = mutate.MediaType(img, types.OCIManifestSchema1)
	img = mutate.ConfigMediaType(img, types.OCIConfigJSON)
	for _, role := range TufRoles {
		layers, err := makeRoleLayers(role, metadata)
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

func makeRoleLayers(role TufRole, tufMetadata *TufMetadata) (*[]mutate.Addendum, error) {
	layers := new([]mutate.Addendum)
	ann := map[string]string{tufRoleAnnotation: ""}
	switch role {
	case metadata.ROOT:
		layers = annotatedMetaLayers(tufMetadata.Root)
	case metadata.SNAPSHOT:
		layers = annotatedMetaLayers(tufMetadata.Snapshot)
	case metadata.TARGETS:
		layers = annotatedMetaLayers(tufMetadata.Targets)
	case metadata.TIMESTAMP:
		ann[tufRoleAnnotation] = fmt.Sprintf("%s.json", role)
		*layers = append(*layers, mutate.Addendum{Layer: static.NewLayer(tufMetadata.Timestamp, tufMetadataMediaType), Annotations: ann})
	default:
		return nil, fmt.Errorf("unsupported TUF role: %s", role)
	}
	return layers, nil
}

func annotatedMetaLayers(meta map[string][]byte) *[]mutate.Addendum {
	layers := new([]mutate.Addendum)
	for name, data := range meta {
		ann := map[string]string{tufRoleAnnotation: name}
		*layers = append(*layers, mutate.Addendum{Layer: static.NewLayer(data, tufMetadataMediaType), Annotations: ann})
	}
	return layers
}

func PushMetadataManifest(imageName string) error {
	metadata, err := GetTufGitRepoMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}
	manifest, err := buildMetadataManifest(metadata)
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
