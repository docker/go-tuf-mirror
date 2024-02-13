package mirror

import "testing"

func TestPushMetadataManifest(t *testing.T) {
	err := PushMetadataManifest()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
