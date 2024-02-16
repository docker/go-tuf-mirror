package tuf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	_ "embed"

	"github.com/docker/go-tuf-mirror/internal/util"
	"github.com/theupdateframework/go-tuf/v2/metadata/config"
	"github.com/theupdateframework/go-tuf/v2/metadata/fetcher"
	"github.com/theupdateframework/go-tuf/v2/metadata/trustedmetadata"
	"github.com/theupdateframework/go-tuf/v2/metadata/updater"
)

type TufClient struct {
	updater *updater.Updater
	cfg     *config.UpdaterConfig
}

// NewTufClient creates a new TUF client
func NewTufClient(initialRoot []byte, tufPath, metadataURL, targetsURL string) (*TufClient, error) {
	tufRootDigest := util.HexHashBytes(initialRoot)

	// create a docker folder for storing tuf stuff
	metadataPath := filepath.Join(tufPath, ".docker", "tuf", tufRootDigest)
	err := os.MkdirAll(metadataPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory '%s': %w", metadataPath, err)
	}
	rootFile := filepath.Join(metadataPath, "root.json")
	var rootBytes []byte
	rootBytes, err = os.ReadFile(rootFile)

	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("failed to read root.json: %w", err)
		}
		// write the root.json file to the metadata directory
		err = os.WriteFile(rootFile, initialRoot, 0644)
		if err != nil {
			return nil, fmt.Errorf("Failed to write root.json %w", err)
		}
		rootBytes = initialRoot
	}
	// readDigest := util.HexHashBytes(rootBytes)
	// if readDigest != tufRootDigest {
	// 	return nil, fmt.Errorf("root.json digest %s doesn't match filesystem: %s at %s", tufRootDigest, readDigest, rootFile)
	// }
	// create updater configuration
	cfg, err := config.New(metadataURL, rootBytes) // default config
	if err != nil {
		return nil, fmt.Errorf("failed to create TUF updater configuration: %w", err)
	}
	cfg.LocalMetadataDir = metadataPath
	cfg.LocalTargetsDir = filepath.Join(metadataPath, "download")
	cfg.RemoteTargetsURL = targetsURL
	//cfg.PrefixTargetsWithHash = true

	// create a new Updater instance
	up, err := updater.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUF updater instance: %w", err)
	}

	// try to build the top-level metadata
	err = up.Refresh()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh trusted metadata: %w", err)
	}

	client := &TufClient{
		updater: up,
		cfg:     cfg,
	}
	return client, nil

}

// DownloadTarget downloads the target file using Updater. The Updater gets the target
// information, verifies if the target is already cached, and if it is not cached,
// downloads the target file.
func (t *TufClient) DownloadTarget(target string, filePath string) (actualFilePath string, data []byte, err error) {
	// search if the desired target is available
	targetInfo, err := t.updater.GetTargetInfo(target)
	if err != nil {
		return "", nil, err
	}

	// target is available, so let's see if the target is already present locally
	actualFilePath, data, err = t.updater.FindCachedTarget(targetInfo, filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed while finding a cached target: %w", err)
	}
	if data != nil {
		return actualFilePath, data, err
	}

	// target is not present locally, so let's try to download it
	actualFilePath, data, err = t.updater.DownloadTarget(targetInfo, filePath, "")
	if err != nil {
		return "", nil, fmt.Errorf("failed to download target file %s - %w", target, err)
	}

	return actualFilePath, data, err
}
func (t *TufClient) GetMetadata() trustedmetadata.TrustedMetadata {
	return t.updater.GetTrustedMetadataSet()
}

func (t *TufClient) MaxRootLength() int64 {
	return t.cfg.RootMaxLength
}

func (t *TufClient) GetPriorRoots(metadataURL string) (map[string][]byte, error) {
	rootMetadata := map[string][]byte{}
	trustedMetadata := t.GetMetadata()
	client := fetcher.DefaultFetcher{}
	for i := 1; i < int(trustedMetadata.Root.Signed.Version); i++ {
		meta, err := client.DownloadFile(metadataURL+fmt.Sprintf("/%d.root.json", i), t.MaxRootLength(), time.Second*15)
		if err != nil {
			return nil, fmt.Errorf("failed to download root metadata: %w", err)
		}
		rootMetadata[fmt.Sprintf("%d.root.json", i)] = meta
	}
	return rootMetadata, nil
}

func (t *TufClient) SetRemoteTargetsURL(url string) {
	t.cfg.RemoteTargetsURL = url
}
