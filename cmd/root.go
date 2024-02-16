package cmd

import (
	_ "embed"
	"fmt"

	"github.com/docker/go-tuf-mirror/pkg/mirror"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	tufPath string
	mirror  *mirror.TufMirror
}

func defaultRootOptions() *rootOptions {
	return &rootOptions{}
}

func newRootCmd(version string) *cobra.Command {
	o := defaultRootOptions()
	cmd := &cobra.Command{
		Use:   "go-tuf-mirror",
		Short: "Mirror TUF metadata to and between OCI registries, filesystems etc",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.PersistentFlags().StringVarP(&o.tufPath, "tuf-path", "t", "", "path on filesystem for tuf root")

	cmd.AddCommand(newMetadataCmd(o))      // metadata subcommand
	cmd.AddCommand(newTargetsCmd(o))       // targets subcommand
	cmd.AddCommand(newVersionCmd(version)) // version subcommand
	cmd.AddCommand(newAllCmd(o))           // all subcommand

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
