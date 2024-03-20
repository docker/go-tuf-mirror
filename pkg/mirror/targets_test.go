package mirror

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/docker/go-tuf-mirror/internal/test"
	"github.com/stretchr/testify/assert"
)

type Layer struct {
	Annotations map[string]string `json:"annotations"`
	Digest      string            `json:"digest"`
}
type Layers struct {
	Layers []Layer `json:"layers"`
}

func TestGetTufTargetsMirror(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	assert.Nil(t, err)

	targets, err := m.GetTufTargetMirrors()
	assert.Nil(t, err)
	assert.Greater(t, len(targets), 0)

	// check for image layer annotations
	for _, target := range targets {
		img := *target.Image
		mf, err := img.RawManifest()
		assert.Nil(t, err)

		// unmarshal manifest with annotations
		l := &Layers{}
		err = json.Unmarshal(mf, l)
		assert.Nil(t, err)

		// check that layers are annotated
		for _, layer := range l.Layers {
			ann, ok := layer.Annotations[tufFileAnnotation]
			assert.True(t, ok)
			parts := strings.Split(ann, ".")
			// <digest>.filename.json
			assert.Equal(t, len(parts), 3)
		}
	}
}

func TestTargetDelegationMetadata(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	tm, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	assert.Nil(t, err)

	targets, err := tm.TufClient.LoadDelegatedTargets("opkl", "targets")
	assert.Nil(t, err)
	assert.Greater(t, len(targets.Signed.Targets), 0)
}

func TestGetDelegatedTargetMirrors(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	assert.Nil(t, err)

	mirrors, err := m.GetDelegatedTargetMirrors()
	assert.Nil(t, err)
	assert.Greater(t, len(mirrors), 0)

	// check for index image annotations
	for _, mirror := range mirrors {
		idx := *mirror.Index
		mf, err := idx.RawManifest()
		assert.Nil(t, err)

		// unmarshal manifest with annotations
		l := &Layers{}
		err = json.Unmarshal(mf, l)
		assert.Nil(t, err)

		// check that layers are annotated
		for _, layer := range l.Layers {
			ann, ok := layer.Annotations[tufFileAnnotation]
			assert.True(t, ok)
			parts := strings.Split(ann, ".")
			// <subdir>/<digest>.filename.json
			assert.Equal(t, len(parts), 3)
		}
	}
}
