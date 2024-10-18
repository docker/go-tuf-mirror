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
	"context"
	_ "embed"
	"fmt"

	"github.com/docker/attest/mirror"
	"github.com/docker/attest/useragent"
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
	ctx := context.Background()
	ctx = useragent.Set(ctx, fmt.Sprintf("go-tuf-mirror/%s (docker)", version))
	if err := newRootCmd(version).ExecuteContext(ctx); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
