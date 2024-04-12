package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-tuf-mirror/internal/util"
	"github.com/docker/go-tuf-mirror/pkg/mirror"
	"github.com/docker/go-tuf-mirror/pkg/types"
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
	cmd.PersistentFlags().StringVarP((&o.metadata), "metadata", "m", mirror.DefaultMetadataURL, fmt.Sprintf("Source metadata location %s<web>, %s<OCI layout>, %s<filesystem> or %s<remote registry>", types.WebPrefix, types.OCIPrefix, types.LocalPrefix, types.RegistryPrefix))
	cmd.PersistentFlags().StringVarP(&o.source, "source", "s", mirror.DefaultMetadataURL, fmt.Sprintf("Source targets location %s<web>, %s<OCI layout>, %s<filesystem> or %s<remote registry>", types.WebPrefix, types.OCIPrefix, types.LocalPrefix, types.RegistryPrefix))
	cmd.PersistentFlags().StringVarP(&o.destination, "destination", "d", "", fmt.Sprintf("Destination targets location %s<OCI layout>, %s<filesystem> or %s<remote registry>", types.OCIPrefix, types.LocalPrefix, types.RegistryPrefix))

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
	if !strings.HasPrefix(o.metadata, types.WebPrefix) && !strings.HasPrefix(o.metadata, types.InsecureWebPrefix) {
		return fmt.Errorf("metadata not implemented: %s", o.source)
	}
	if !strings.HasPrefix(o.source, types.WebPrefix) && !strings.HasPrefix(o.source, types.InsecureWebPrefix) {
		return fmt.Errorf("source not implemented: %s", o.source)
	}
	if !(strings.HasPrefix(o.destination, types.RegistryPrefix) || strings.HasPrefix(o.destination, types.OCIPrefix)) {
		return fmt.Errorf("destination not implemented: %s", o.destination)
	}
	if !util.IsValidUrl(o.source) {
		return fmt.Errorf("invalid source url: %s", o.source)
	}
	if strings.HasPrefix(o.destination, types.RegistryPrefix) && strings.Contains(strings.TrimPrefix(o.destination, types.RegistryPrefix), ":") {
		return fmt.Errorf("destination registry should not specify tag: %s", o.destination)
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
		m, err = mirror.NewTufMirror(o.rootOptions.tufRootBytes, tufPath, o.metadata, o.source)
		if err != nil {
			return fmt.Errorf("failed to create TUF mirror: %w", err)
		}
	} else {
		// set remote targets url for existing mirror
		m.TufClient.SetRemoteTargetsURL(o.source)
	}

	// create target manifests
	targets, err := m.GetTufTargetMirrors()
	if err != nil {
		return fmt.Errorf("failed to create target mirrors: %w", err)
	}

	// create delegated target manifests
	var delegated []*mirror.MirrorIndex
	if o.rootOptions.full {
		delegated, err = m.GetDelegatedTargetMirrors()
		if err != nil {
			return fmt.Errorf("failed to create delegated target index manifests: %w", err)
		}
	}

	// save target manifests
	switch {
	case strings.HasPrefix(o.destination, types.OCIPrefix):
		outputPath := strings.TrimPrefix(o.destination, types.OCIPrefix)
		for _, t := range targets {
			path := filepath.Join(outputPath, t.Tag)
			err = mirror.SaveAsOCILayout(t.Image, path)
			if err != nil {
				return fmt.Errorf("failed to save target as OCI layout: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Target manifest layout saved to %s\n", path)
		}
		for _, d := range delegated {
			path := filepath.Join(outputPath, d.Tag)
			err = mirror.SaveAsOCILayout(d.Index, path)
			if err != nil {
				return fmt.Errorf("failed to save delegated target index as OCI layout: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Delegated target index manifest layout saved to %s\n", path)
		}
	case strings.HasPrefix(o.destination, types.RegistryPrefix):
		repo := strings.TrimPrefix(o.destination, types.RegistryPrefix)
		for _, t := range targets {
			imageName := fmt.Sprintf("%s:%s", repo, t.Tag)
			err = mirror.PushToRegistry(t.Image, imageName)
			if err != nil {
				return fmt.Errorf("failed to push target manifest: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Target manifest pushed to %s\n", imageName)
		}
		for _, d := range delegated {
			imageName := fmt.Sprintf("%s:%s", repo, d.Tag)
			err = mirror.PushToRegistry(d.Index, imageName)
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
