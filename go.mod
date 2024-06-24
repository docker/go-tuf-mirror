module github.com/docker/go-tuf-mirror

go 1.22.2

require (
	github.com/docker/attest v0.1.5
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.15.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/cli v26.1.3+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker v27.0.0+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.1 // indirect
	github.com/go-jose/go-jose/v4 v4.0.1 // indirect
	github.com/google/go-containerregistry v0.19.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/letsencrypt/boulder v0.0.0-20240515153123-6ae6aa8e9055 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.8.0 // indirect
	github.com/sigstore/sigstore v1.8.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/theupdateframework/go-tuf/v2 v2.0.0-20240504210453-5a634eb214ae // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	github.com/vbatts/tar-split v0.11.5 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/term v0.20.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/theupdateframework/go-tuf/v2 => github.com/mrjoelkamp/go-tuf/v2 v2.0.1 // for https://github.com/theupdateframework/go-tuf/pull/632
