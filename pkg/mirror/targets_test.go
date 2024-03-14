package mirror

import (
	"testing"

	"github.com/docker/go-tuf-mirror/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestGetTufTargetsMirror(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := m.GetTufTargetMirrors()
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) == 0 {
		t.Error("Expected non-empty targets")
	}
}

func TestTargetDelegationMetadata(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	tm, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := tm.TufClient.LoadDelegatedTargets("opkl", "targets")
	if err != nil {
		t.Fatal(err)
	}
	assert.Greater(t, len(targets.Signed.Targets), 0)
}

func TestGetDelegatedTargetMirrors(t *testing.T) {
	path := test.CreateTempDir(t, "tuf_temp")
	m, err := NewTufMirror(path, DefaultMetadataURL, DefaultTargetsURL)
	if err != nil {
		t.Fatal(err)
	}
	mirrors, err := m.GetDelegatedTargetMirrors()
	if err != nil {
		t.Fatal(err)
	}
	if len(mirrors) == 0 {
		t.Error("Expected non-empty targets")
	}
}
