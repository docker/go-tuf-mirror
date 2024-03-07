package mirror

import (
	"encoding/json"
	"testing"

	"github.com/docker/go-tuf-mirror/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestGetTufMetadataMirror(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	if err != nil {
		t.Fatal(err)
	}
	tufMetadata, err := m.getTufMetadataMirror(DefaultMetadataURL)
	if err != nil {
		t.Fatal(err)
	}
	if len(tufMetadata.Root) == 0 {
		t.Error("Expected non-empty root metadata")
	}
	if len(tufMetadata.Snapshot) == 0 {
		t.Error("Expected non-empty snapshot metadata")
	}
	if len(tufMetadata.Targets) == 0 {
		t.Error("Expected non-empty targets metadata")
	}
	if len(tufMetadata.Timestamp) == 0 {
		t.Error("Expected non-empty timestamp metadata")
	}
}

func TestGetMetadataManifest(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	if err != nil {
		t.Fatal(err)
	}
	img, err := m.GetMetadataManifest(DefaultMetadataURL)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, img)
	image := *img
	mf, err := image.RawManifest()
	if err != nil {
		t.Fatal(err)
	}
	type Annotations struct {
		Annotations map[string]string `json:"annotations"`
	}
	type Layers struct {
		Layers []Annotations `json:"layers"`
	}
	l := &Layers{}
	err = json.Unmarshal(mf, l)
	if err != nil {
		t.Fatal(err)
	}
	for _, layer := range l.Layers {
		_, ok := layer.Annotations[tufFileAnnotation]
		if !ok {
			t.Fatalf("missing annotations")
		}
	}
}

func TestGetDelegatedMetadataMirrors(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	if err != nil {
		t.Fatal(err)
	}
	delegations, err := m.GetDelegatedMetadataMirrors()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, delegations)
	assert.Greater(t, len(delegations), 0)
}
