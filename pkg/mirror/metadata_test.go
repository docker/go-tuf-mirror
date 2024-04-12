package mirror

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/go-tuf-mirror/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/theupdateframework/go-tuf/v2/metadata"
)

func TestGetTufMetadataMirror(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join("..", "..", "internal", "tuf", "testdata", "test-repo"))))
	defer server.Close()

	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(DevRoot, path, server.URL+"/metadata", server.URL+"/targets")
	assert.Nil(t, err)

	tufMetadata, err := m.getTufMetadataMirror(server.URL + "/metadata")
	assert.Nil(t, err)

	// check that all roles are not empty
	assert.Greater(t, len(tufMetadata.Root), 0)
	assert.Greater(t, len(tufMetadata.Snapshot), 0)
	assert.Greater(t, len(tufMetadata.Targets), 0)
	assert.Greater(t, len(tufMetadata.Timestamp), 0)
}

func TestGetMetadataManifest(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join("..", "..", "internal", "tuf", "testdata", "test-repo"))))
	defer server.Close()

	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(DevRoot, path, server.URL+"/metadata", server.URL+"/targets")
	assert.Nil(t, err)

	img, err := m.GetMetadataManifest(server.URL + "/metadata")
	assert.Nil(t, err)
	assert.NotNil(t, img)

	image := *img
	mf, err := image.RawManifest()
	assert.Nil(t, err)

	type Annotations struct {
		Annotations map[string]string `json:"annotations"`
	}
	type Layers struct {
		Layers []Annotations `json:"layers"`
	}
	l := &Layers{}
	err = json.Unmarshal(mf, l)
	assert.Nil(t, err)

	// check that layers are annotated and use consistent snapshot naming
	for _, layer := range l.Layers {
		ann, ok := layer.Annotations[tufFileAnnotation]
		assert.True(t, ok)
		// check for consistent snapshot version
		parts := strings.Split(ann, ".")
		if parts[0] == metadata.TIMESTAMP {
			continue
		}
		_, err := strconv.Atoi(parts[0])
		assert.Nil(t, err)
	}
}

func TestGetDelegatedMetadataMirrors(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir(filepath.Join("..", "..", "internal", "tuf", "testdata", "test-repo"))))
	defer server.Close()

	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(DevRoot, path, server.URL+"/metadata", server.URL+"/targets")
	assert.Nil(t, err)

	delegations, err := m.GetDelegatedMetadataMirrors()
	assert.Nil(t, err)

	assert.NotNil(t, delegations)
	assert.Greater(t, len(delegations), 0)
}
