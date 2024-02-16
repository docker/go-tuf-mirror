package mirror

import (
	"encoding/json"
	"testing"
)

func TestGetTufGitRepoMetadata(t *testing.T) {
	tufMetadata, err := GetTufGitRepoMetadata(DefaultMetadataURL)
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

func TestCreateMetadataManifest(t *testing.T) {
	img, err := CreateMetadataManifest(DefaultMetadataURL)
	if err != nil {
		t.Fatal(err)
	}
	if *img == nil {
		t.Error("Expected non-nil image")
	}
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
		_, ok := layer.Annotations[tufRoleAnnotation]
		if !ok {
			t.Fatalf("missing annotations")
		}
	}
}
