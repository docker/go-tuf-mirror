package mirror

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/theupdateframework/go-tuf/v2/metadata"
)

func (m *TufMirror) GetTufTargetMirrors() ([]*MirrorImage, error) {
	targetMirrors := []*MirrorImage{}
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
		targetMirrors = append(targetMirrors, &MirrorImage{Image: &img, Tag: name})
	}
	return targetMirrors, nil
}
