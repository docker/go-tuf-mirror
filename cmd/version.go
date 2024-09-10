package cmd

import (
	"fmt"

	av "github.com/docker/attest/version"
	"github.com/spf13/cobra"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "go-tuf-mirror version",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if version == "" {
				version = "unknown"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "github.com/docker/go-tuf-mirror: %s\n", version)
			attestVersion, err := av.Get()
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "github.com/docker/attest: unknown")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "github.com/docker/attest: %s\n", attestVersion)
			}
		},
	}
}
