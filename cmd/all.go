package cmd

import (
	"fmt"
	"log"

	"github.com/docker/go-tuf-mirror/pkg/mirror"
	"github.com/docker/go-tuf-mirror/pkg/types"
	"github.com/spf13/cobra"
)

type allOptions struct {
	srcMeta     string
	dstMeta     string
	srcTargets  string
	dstTargets  string
	rootOptions *rootOptions
}

func defaultAllOptions(opts *rootOptions) *allOptions {
	return &allOptions{
		rootOptions: opts,
	}
}

func newAllCmd(opts *rootOptions) *cobra.Command {
	o := defaultAllOptions(opts)

	cmd := &cobra.Command{
		Use:          "all",
		Short:        "Mirror TUF metadata and targets to and between OCI registries, filesystems etc",
		SilenceUsage: false,
		RunE:         o.run,
	}
	cmd.Flags().StringVar(&o.srcMeta, "source-metadata", mirror.DefaultMetadataURL, fmt.Sprintf("Source metadata location %s<web>, %s<OCI layout>, %s<filesystem> or %s<remote registry>", types.WebPrefix, types.OCIPrefix, types.LocalPrefix, types.RegistryPrefix))
	cmd.Flags().StringVar(&o.dstMeta, "dest-metadata", "", fmt.Sprintf("Destination metadata location %s<OCI layout>, %s<filesystem> or %s<remote registry>", types.OCIPrefix, types.LocalPrefix, types.RegistryPrefix))
	cmd.Flags().StringVar(&o.srcTargets, "source-targets", mirror.DefaultTargetsURL, fmt.Sprintf("Source targets location %s<web>, %s<OCI layout>, %s<filesystem> or %s<remote registry>", types.WebPrefix, types.OCIPrefix, types.LocalPrefix, types.RegistryPrefix))
	cmd.Flags().StringVar(&o.dstTargets, "dest-targets", "", fmt.Sprintf("Destination targets location %s<OCI layout>, %s<filesystem> or %s<remote registry>", types.OCIPrefix, types.LocalPrefix, types.RegistryPrefix))

	err := cmd.MarkFlagRequired("source-metadata")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	err = cmd.MarkFlagRequired("dest-metadata")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	err = cmd.MarkFlagRequired("source-targets")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	err = cmd.MarkFlagRequired("dest-targets")
	if err != nil {
		log.Fatalf("failed to mark flag required: %s", err)
	}
	return cmd
}

func (o *allOptions) run(cmd *cobra.Command, args []string) error {
	metadata := newMetadataCmd(o.rootOptions)
	metadata.SetOut(cmd.OutOrStdout())
	targets := newTargetsCmd(o.rootOptions)
	targets.SetOut(cmd.OutOrStdout())

	_ = metadata.PersistentFlags().Set("source", o.srcMeta)
	_ = metadata.PersistentFlags().Set("destination", o.dstMeta)
	_ = targets.PersistentFlags().Set("source", o.srcTargets)
	_ = targets.PersistentFlags().Set("destination", o.dstTargets)
	_ = targets.PersistentFlags().Set("metadata", o.srcMeta)

	err := metadata.Execute()
	if err != nil {
		return fmt.Errorf("error mirroring metadata: %w", err)
	}
	err = targets.Execute()
	if err != nil {
		return fmt.Errorf("error mirroring targets: %w", err)
	}
	return nil
}
