package mirror

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-tuf-mirror/internal/tuf"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/theupdateframework/go-tuf/v2/metadata"
)

//go:embed 1.root-staging.json
var InitialRoot []byte

const (
	DefaultMetadataURL   = "https://docker.github.io/tuf-staging/metadata"
	DefaultTargetsURL    = "https://docker.github.io/tuf-staging/targets"
	tufMetadataMediaType = "application/vnd.tuf.metadata+json"
	tufTargetMediaType   = "application/vnd.tuf.target"
	tufFileAnnotation    = "tuf.io/filename"
)

type TufRole string

var TufRoles = []TufRole{metadata.ROOT, metadata.SNAPSHOT, metadata.TARGETS, metadata.TIMESTAMP}

type TufMetadataMirror struct {
	Root      map[string][]byte
	Snapshot  map[string][]byte
	Targets   map[string][]byte
	Timestamp []byte
}

type TufTargetMirror struct {
	Image *v1.Image
	Tag   string
}

type TufMirror struct {
	TufClient   *tuf.TufClient
	tufPath     string
	metadataURL string
	targetsURL  string
}

func NewTufMirror(tufPath string, metadataURL string, targetsURL string) (*TufMirror, error) {
	tufClient, err := tuf.NewTufClient(InitialRoot, tufPath, metadataURL, targetsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUF client: %w", err)
	}
	return &TufMirror{TufClient: tufClient, tufPath: tufPath, metadataURL: metadataURL, targetsURL: targetsURL}, nil
}

func (m *TufMirror) getTufMetadataMirror(metadataURL string) (*TufMetadataMirror, error) {
	trustedMetadata := m.TufClient.GetMetadata()

	rootMetadata := map[string][]byte{}
	rootVersion := trustedMetadata.Root.Signed.Version
	// get the previous versions of root metadata if any
	if rootVersion != 1 {
		var err error
		rootMetadata, err = m.TufClient.GetPriorRoots(metadataURL)
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
	return &TufMetadataMirror{
		Root:      rootMetadata,
		Snapshot:  map[string][]byte{snapshotName: snapshotBytes},
		Targets:   map[string][]byte{targetsName: targetsBytes},
		Timestamp: timestampBytes,
	}, nil
}

func (m *TufMirror) GetTufTargetMirrors() ([]*TufTargetMirror, error) {
	targetMirrors := []*TufTargetMirror{}
	md := m.TufClient.GetMetadata()
	targets := md.Targets[metadata.TARGETS].Signed.Targets
	for _, t := range targets {
		_, data, err := m.TufClient.DownloadTarget(t.Path, filepath.Join(m.tufPath, "download"))
		if err != nil {
			return nil, fmt.Errorf("failed to download target %s: %w", t.Path, err)
		}
		img := empty.Image
		img = mutate.MediaType(img, types.OCIManifestSchema1)
		img = mutate.ConfigMediaType(img, types.OCIConfigJSON)
		hash, ok := t.Hashes["sha256"]
		if !ok {
			return nil, fmt.Errorf("missing sha256 hash for target %s", t.Path)
		}
		name := strings.Join([]string{hash.String(), t.Path}, ".")
		ann := map[string]string{tufFileAnnotation: name}
		layer := mutate.Addendum{Layer: static.NewLayer(data, tufTargetMediaType), Annotations: ann}
		img, err = mutate.Append(img, layer)
		if err != nil {
			return nil, fmt.Errorf("failed to append role layer to image: %w", err)
		}
		targetMirrors = append(targetMirrors, &TufTargetMirror{Image: &img, Tag: name})
	}
	return targetMirrors, nil
}

func (m *TufMirror) buildMetadataManifest(metadata *TufMetadataMirror) (*v1.Image, error) {
	img := empty.Image
	img = mutate.MediaType(img, types.OCIManifestSchema1)
	img = mutate.ConfigMediaType(img, types.OCIConfigJSON)
	for _, role := range TufRoles {
		layers, err := m.makeRoleLayers(role, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to make role layer: %w", err)
		}
		img, err = mutate.Append(img, *layers...)
		if err != nil {
			return nil, fmt.Errorf("failed to append role layer to image: %w", err)
		}
	}
	return &img, nil
}

func (m *TufMirror) makeRoleLayers(role TufRole, tufMetadata *TufMetadataMirror) (*[]mutate.Addendum, error) {
	layers := new([]mutate.Addendum)
	ann := map[string]string{tufFileAnnotation: ""}
	switch role {
	case metadata.ROOT:
		layers = m.annotatedMetaLayers(tufMetadata.Root)
	case metadata.SNAPSHOT:
		layers = m.annotatedMetaLayers(tufMetadata.Snapshot)
	case metadata.TARGETS:
		layers = m.annotatedMetaLayers(tufMetadata.Targets)
	case metadata.TIMESTAMP:
		ann[tufFileAnnotation] = fmt.Sprintf("%s.json", role)
		*layers = append(*layers, mutate.Addendum{Layer: static.NewLayer(tufMetadata.Timestamp, tufMetadataMediaType), Annotations: ann})
	default:
		return nil, fmt.Errorf("unsupported TUF role: %s", role)
	}
	return layers, nil
}

func (m *TufMirror) annotatedMetaLayers(meta map[string][]byte) *[]mutate.Addendum {
	layers := new([]mutate.Addendum)
	for name, data := range meta {
		ann := map[string]string{tufFileAnnotation: name}
		*layers = append(*layers, mutate.Addendum{Layer: static.NewLayer(data, tufMetadataMediaType), Annotations: ann})
	}
	return layers
}

func (m *TufMirror) CreateMetadataManifest(metadataURL string) (*v1.Image, error) {
	metadata, err := m.getTufMetadataMirror(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	manifest, err := m.buildMetadataManifest(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to build metadata manifest: %w", err)
	}
	return manifest, nil
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
