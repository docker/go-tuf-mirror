package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/attest/mirror"
	"github.com/docker/attest/oci"
	"github.com/docker/attest/tuf"
	"github.com/docker/go-tuf-mirror/internal/util"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
)

type targetsOptions struct {
	source      string
	destination string
	metadata    string
	rootOptions *rootOptions
}

func defaultTargetsOptions(opts *rootOptions) *targetsOptions {
	return &targetsOptions{
		rootOptions: opts,
	}
}

func newTargetsCmd(opts *rootOptions) *cobra.Command {
	o := defaultTargetsOptions(opts)

	cmd := &cobra.Command{
		Use:          "targets",
		Short:        "Mirror TUF targets to and between OCI registries, filesystems etc",
		SilenceUsage: false,
		RunE:         o.run,
	}
	cmd.PersistentFlags().StringVarP((&o.metadata), "metadata", "m", mirror.DefaultMetadataURL, fmt.Sprintf("Source metadata location %s<web>, %s<OCI layout>, %s<filesystem> or %s<remote registry>", WebPrefix, OCIPrefix, LocalPrefix, RegistryPrefix))
	cmd.PersistentFlags().StringVarP(&o.source, "source", "s", mirror.DefaultMetadataURL, fmt.Sprintf("Source targets location %s<web>, %s<OCI layout>, %s<filesystem> or %s<remote registry>", WebPrefix, OCIPrefix, LocalPrefix, RegistryPrefix))
	cmd.PersistentFlags().StringVarP(&o.destination, "destination", "d", "", fmt.Sprintf("Destination targets location %s<OCI layout>, %s<filesystem> or %s<remote registry>", OCIPrefix, LocalPrefix, RegistryPrefix))

	err := cmd.MarkPersistentFlagRequired("metadata")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	err = cmd.MarkPersistentFlagRequired("source")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	err = cmd.MarkPersistentFlagRequired("destination")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	return cmd
}

func (o *targetsOptions) run(cmd *cobra.Command, args []string) error {
	// only support web to registry or oci layout for now
	if !strings.HasPrefix(o.metadata, WebPrefix) && !strings.HasPrefix(o.metadata, InsecureWebPrefix) {
		return fmt.Errorf("metadata not implemented: %s", o.source)
	}
	if !strings.HasPrefix(o.source, WebPrefix) && !strings.HasPrefix(o.source, InsecureWebPrefix) {
		return fmt.Errorf("source not implemented: %s", o.source)
	}
	if !(strings.HasPrefix(o.destination, RegistryPrefix) || strings.HasPrefix(o.destination, OCIPrefix)) {
		return fmt.Errorf("destination not implemented: %s", o.destination)
	}
	if !util.IsValidUrl(o.source) {
		return fmt.Errorf("invalid source url: %s", o.source)
	}
	if strings.HasPrefix(o.destination, RegistryPrefix) {
		_, err := name.NewRepository(strings.TrimPrefix(o.destination, RegistryPrefix))
		if err != nil {
			return fmt.Errorf("failed to parse destination registry reference: %w", err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Mirroring TUF targets %s to %s\n", o.source, o.destination)

	// use existing mirror from root or create new one
	m := o.rootOptions.mirror
	if m == nil {
		var tufPath string
		var err error
		if o.rootOptions.tufPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			tufPath = filepath.Join(home, ".docker", "tuf")
		} else {
			tufPath = strings.TrimSpace(o.rootOptions.tufPath)
		}
		root, err := tuf.GetEmbeddedRoot(o.rootOptions.tufRoot)
		if err != nil {
			return fmt.Errorf("failed to get root bytes: %w", err)
		}
		m, err = mirror.NewTUFMirror(cmd.Context(), root.Data, tufPath, o.metadata, o.source, tuf.NewDefaultVersionChecker())
		if err != nil {
			return fmt.Errorf("failed to create TUF mirror: %w", err)
		}
	} else {
		// set remote targets url for existing mirror
		m.TUFClient.SetRemoteTargetsURL(o.source)
	}

	// create target manifests
	targets, err := m.GetTUFTargetMirrors()
	if err != nil {
		return fmt.Errorf("failed to create target mirrors: %w", err)
	}

	// create delegated target manifests
	var delegated []*mirror.Index
	if o.rootOptions.full {
		delegated, err = m.GetDelegatedTargetMirrors()
		if err != nil {
			return fmt.Errorf("failed to create delegated target index manifests: %w", err)
		}
	}

	// save target manifests
	switch {
	case strings.HasPrefix(o.destination, OCIPrefix):
		outputPath := strings.TrimPrefix(o.destination, OCIPrefix)
		for _, t := range targets {
			path := filepath.Join(outputPath, t.Tag)
			err = oci.SaveImageAsOCILayout(t.Image, path)
			if err != nil {
				return fmt.Errorf("failed to save target as OCI layout: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Target manifest layout saved to %s\n", path)
		}
		for _, d := range delegated {
			path := filepath.Join(outputPath, d.Tag)
			err = oci.SaveIndexAsOCILayout(d.Index, path)
			if err != nil {
				return fmt.Errorf("failed to save delegated target index as OCI layout: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Delegated target index manifest layout saved to %s\n", path)
		}
	case strings.HasPrefix(o.destination, RegistryPrefix):
		repo := strings.TrimPrefix(o.destination, RegistryPrefix)
		for _, t := range targets {
			imageName := fmt.Sprintf("%s:%s", repo, t.Tag)
			err = oci.PushImageToRegistry(cmd.Context(), t.Image, imageName)
			if err != nil {
				return fmt.Errorf("failed to push target manifest: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Target manifest pushed to %s\n", imageName)
		}
		for _, d := range delegated {
			imageName := fmt.Sprintf("%s:%s", repo, d.Tag)
			err = oci.PushIndexToRegistry(cmd.Context(), d.Index, imageName)
			if err != nil {
				return fmt.Errorf("failed to push delegated target index manifest: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Delegated target index manifest pushed to %s\n", imageName)
		}
	default:
		return fmt.Errorf("destination not implemented: %s", o.destination)
	}
	return nil
}
