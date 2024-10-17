package cmd

import (
	"fmt"

	"github.com/docker/attest/version"
	"github.com/spf13/cobra"
)

func newVersionCmd(v string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "go-tuf-mirror version",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if v == "" {
				v = "unknown"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "github.com/docker/go-tuf-mirror: %s\n", v)
			fetcher := version.NewGoVersionFetcher()
			attestVersion, err := fetcher.Get()
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "github.com/docker/attest: unknown")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "github.com/docker/attest: %s\n", attestVersion)
			}
		},
	}
}
