package cmd

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/docker/attest/pkg/mirror"
	"github.com/docker/go-tuf-mirror/internal/embed"
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
	tufPath      string
	tufRootBytes []byte
	mirror       *mirror.TufMirror
	full         bool
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
	root := cmd.PersistentFlags().StringP("tuf-root", "r", "", "specify embedded tuf root [dev, staging], default [staging]")
	switch *root {
	case "dev":
		o.tufRootBytes = embed.DevRoot
	case "staging":
		o.tufRootBytes = embed.StagingRoot
	case "":
		o.tufRootBytes = embed.DefaultRoot
	default:
		log.Fatalf("invalid tuf root: %s", *root)
	}

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
