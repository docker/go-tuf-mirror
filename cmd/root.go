package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

type rootOptions struct {
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

	cmd.AddCommand(newMetadataCmd(o))      // metadata subcommand
	cmd.AddCommand(newVersionCmd(version)) // version subcommand

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
