package cmd

import (
	_ "embed"
	"fmt"

	"github.com/docker/attest/mirror"
	"github.com/spf13/cobra"
)

const (
	OCIPrefix         = "oci://"    // filesystem oci layout
	RegistryPrefix    = "docker://" // remote registry
	LocalPrefix       = "file://"   // local filesystem
	WebPrefix         = "https://"  // web
	InsecureWebPrefix = "http://"   // insecure web
)

type rootOptions struct {
	tufPath string
	tufRoot string
	mirror  *mirror.TUFMirror
	full    bool
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
	cmd.PersistentFlags().BoolVarP(&o.full, "full", "f", false, "Mirror full metadata/targets (includes delegated targets)")
	cmd.PersistentFlags().StringVarP(&o.tufRoot, "tuf-root", "r", "", "specify embedded tuf root [dev, staging, prod], default [prod]")

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
