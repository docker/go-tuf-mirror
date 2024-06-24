package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/attest/pkg/mirror"
	"github.com/docker/attest/pkg/tuf"
	"github.com/docker/go-tuf-mirror/internal/util"
	"github.com/spf13/cobra"
)

type metadataOptions struct {
	source      string
	destination string
	rootOptions *rootOptions
}

func defaultMetadataOptions(opts *rootOptions) *metadataOptions {
	return &metadataOptions{
		rootOptions: opts,
	}
}

func newMetadataCmd(opts *rootOptions) *cobra.Command {
	o := defaultMetadataOptions(opts)

	cmd := &cobra.Command{
		Use:          "metadata",
		Short:        "Mirror TUF metadata to and between OCI registries, filesystems etc",
		SilenceUsage: false,
		RunE:         o.run,
	}
	cmd.PersistentFlags().StringVarP(&o.source, "source", "s", mirror.DefaultMetadataURL, fmt.Sprintf("Source metadata location %s<web>, %s<OCI layout>, %s<filesystem> or %s<remote registry>", WebPrefix, OCIPrefix, LocalPrefix, RegistryPrefix))
	cmd.PersistentFlags().StringVarP(&o.destination, "destination", "d", "", fmt.Sprintf("Destination metadata location %s<OCI layout>, %s<filesystem> or %s<remote registry>", OCIPrefix, LocalPrefix, RegistryPrefix))

	err := cmd.MarkPersistentFlagRequired("source")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	err = cmd.MarkPersistentFlagRequired("destination")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	return cmd
}

func (o *metadataOptions) run(cmd *cobra.Command, args []string) error {
	// only support web to registry or oci layout for now
	if !strings.HasPrefix(o.source, WebPrefix) && !strings.HasPrefix(o.source, InsecureWebPrefix) {
		return fmt.Errorf("source not implemented: %s", o.source)
	}
	if !(strings.HasPrefix(o.destination, RegistryPrefix) || strings.HasPrefix(o.destination, OCIPrefix)) {
		return fmt.Errorf("destination not implemented: %s", o.destination)
	}
	if !util.IsValidUrl(o.source) {
		return fmt.Errorf("invalid source url: %s", o.source)
	}
	var tufPath string
	if o.rootOptions.tufPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		tufPath = filepath.Join(home, ".docker", "tuf")
	} else {
		tufPath = strings.TrimSpace(o.rootOptions.tufPath)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Mirroring TUF metadata %s to %s\n", o.source, o.destination)
	m, err := mirror.NewTufMirror(o.rootOptions.tufRootBytes, tufPath, o.source, "", tuf.NewVersionChecker())
	if err != nil {
		return fmt.Errorf("failed to create TUF mirror: %w", err)
	}
	// set mirror in root options for reuse in targets
	o.rootOptions.mirror = m

	// create metadata image
	image, err := m.GetMetadataManifest(o.source)
	if err != nil {
		return fmt.Errorf("failed to create metadata manifest: %w", err)
	}

	// create delegated metadata manifests
	var delegated []*mirror.MirrorImage
	if o.rootOptions.full {
		delegated, err = m.GetDelegatedMetadataMirrors()
		if err != nil {
			return fmt.Errorf("failed to create delegated metadata manifests: %w", err)
		}
	}

	// save metadata manifest
	switch {
	case strings.HasPrefix(o.destination, OCIPrefix):
		path := strings.TrimPrefix(o.destination, OCIPrefix)
		err = mirror.SaveImageAsOCILayout(image, path)
		if err != nil {
			return fmt.Errorf("failed to save metadata as OCI layout: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Metadata manifest layout saved to %s\n", path)
		for _, d := range delegated {
			path := filepath.Join(path, d.Tag)
			err = mirror.SaveImageAsOCILayout(d.Image, path)
			if err != nil {
				return fmt.Errorf("failed to save delegated metadata as OCI layout: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Delegated metadata manifest layout saved to %s\n", path)
		}
	case strings.HasPrefix(o.destination, RegistryPrefix):
		imageName := strings.TrimPrefix(o.destination, RegistryPrefix)
		err = mirror.SaveImageAsOCILayout(image, imageName)
		if err != nil {
			return fmt.Errorf("failed to push metadata manifest: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Metadata manifest pushed to %s\n", imageName)
		for _, d := range delegated {
			repo, _, ok := strings.Cut(imageName, ":")
			if !ok {
				return fmt.Errorf("failed to get repo from image name: %s", imageName)
			}
			imageName := fmt.Sprintf("%s:%s", repo, d.Tag)
			err = mirror.PushImageToRegistry(d.Image, imageName)
			if err != nil {
				return fmt.Errorf("failed to push delegated metadata manifest: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Delegated metadata manifest pushed to %s\n", imageName)
		}
	}
	return nil
}
