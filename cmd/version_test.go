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
	"bytes"
	"fmt"
	"io"
	"testing"

	av "github.com/docker/attest/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	version := "v1.0.0"
	cmd := newVersionCmd(version)
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	err := cmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(b)
	require.NoError(t, err)
	fetcher := av.NewGoVersionFetcher()
	attestVersion, err := fetcher.Get()
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("github.com/docker/go-tuf-mirror: %s\ngithub.com/docker/attest: %s\n", version, attestVersion), string(out))
}
