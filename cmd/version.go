/*
   Copyright Docker go-tuf-mirror authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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
