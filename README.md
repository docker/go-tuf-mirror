# go-tuf-mirror

Mirror TUF metadata to/between OCI registries

<div align="left">
<img src="https://github.com/docker/go-tuf-mirror/actions/workflows/test.yml/badge.svg" alt="drawing"/>
</div>

## Usage

### GitHub Actions

Example GHA workflow:

```yaml
name: Run go-tuf-mirror
on:
  workflow_dispatch:
jobs:
  mirror:
    runs-on: ubuntu-latest
    env:
      DOCKER_CONFIG: ${{ github.workspace }}/.docker
    steps:
      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: dockerpublicbot
          password: ${{ secrets.DOCKERPUBLICBOT_WRITE_PAT }}
      - name: Mirror metadata
        uses: docker/go-tuf-mirror/actions/metadata@v0.1.0
        with:
          source: https://docker.github.io/tuf-staging/metadata
          destination: docker://docker/tuf-metadata:latest
      - name: Mirror targets
        uses: docker/go-tuf-mirror/actions/targets@v0.1.0
        with:
          metadata: https://docker.github.io/tuf-staging/metadata
          source: https://docker.github.io/tuf-staging/targets
          destination: docker://docker/tuf-targets
```

### Mirror only metadata from web

1. Build `go-tuf-mirror`
   ```sh
   make build
   ```
1. Run `metadata` command

   ```sh
   ./go-tuf-mirror metadata -s <metadata location> -d <metadata output location>
   ```

   example:

   ```sh
   # output metadata to docker registry
   ./go-tuf-mirror metadata -s https://docker.github.io/tuf-staging/metadata -d docker://docker/tuf-metadata:latest

   Mirroring TUF metadata https://docker.github.io/tuf-staging/metadata to docker://docker/tuf-metadata:latest
   Metadata manifest pushed to docker/tuf-metadata:latest
   ```

#### Mirror delegated targets metadata

1. Run `metadata` command with the `-f` flag

   example:

   ```sh
   ./go-tuf-mirror metadata -f -s "https://docker.github.io/tuf-staging/metadata" -d "docker://docker/tuf-metadata:latest"

   Mirroring TUF metadata https://docker.github.io/tuf-staging/metadata to docker://docker/tuf-metadata:latest
   Metadata manifest pushed to docker/tuf-metadata:latest
   Delegated metadata manifest pushed to docker/tuf-metadata:opkl
   Delegated metadata manifest pushed to docker/tuf-metadata:doi
   ```

### Mirror only targets from web

1. Build `go-tuf-mirror`
   ```sh
   make build
   ```
1. Run `metadata` command

   ```sh
   ./go-tuf-mirror targets -m <source metadata location> -s <source targets location>  -d <destination targets location>
   ```

   example:

   ```sh
   # output targets to docker registry
   ./go-tuf-mirror targets -m https://docker.github.io/tuf-staging/metadata -s https://docker.github.io/tuf-staging/targets  -d docker://docker/tuf-targets

   Mirroring TUF targets https://docker.github.io/tuf-staging/targets to docker://docker/tuf-targets
   Target manifest pushed to docker/tuf-targets:ecc736303caf8cf22ef00df2db3c411a563030c2e1e7ae24f4e38113e7ad610d.doi-signing-stage.pem
   Target manifest pushed to docker/tuf-targets:3965bb0a873cff50e16b277444d659553ab79c9632a1fb03a6d9360af536c142.image-signer-verifier.pem
   Target manifest pushed to docker/tuf-targets:e4dc114275694612ee236b231990d606b7879d05f64809611545c8234efb6cd4.doi-signing-key.pem
   Target manifest pushed to docker/tuf-targets:5ddbaf12a091d0b877b7574af7cc19bf85023d649a520ccfebc0f2b5f8c2c4de.doi-signing-prod.pem
   ```

### Mirror metadata and targets from web

1. Build `go-tuf-mirror`

   ```sh
   make build
   ```

1. Run `all` command

   ```sh
   ./go-tuf-mirror all --source-metadata <metadata location> --source-targets <targets location> --dest-metadata <metadata output location> --dest-targets <targets output location>
   ```

   example:

   ```sh
   # outputs metadata and targets to local OCI layout
   ./go-tuf-mirror all --source-metadata "https://docker.github.io/tuf-staging/metadata" --source-targets "https://docker.github.io/tuf-staging/targets" --dest-targets "oci://./tmp/targets" --dest-metadata "oci://./tmp/metadata"

   Mirroring TUF metadata https://docker.github.io/tuf-staging/metadata to oci://./tmp/metadata
   Metadata manifest layout saved to ./tmp/metadata

   Mirroring TUF targets https://docker.github.io/tuf-staging/targets to oci://./tmp/targets
   Target manifest layout saved to tmp/targets/ecc736303caf8cf22ef00df2db3c411a563030c2e1e7ae24f4e38113e7ad610d.doi-signing-stage.pem
   Target manifest layout saved to tmp/targets/3965bb0a873cff50e16b277444d659553ab79c9632a1fb03a6d9360af536c142.image-signer-verifier.pem
   Target manifest layout saved to tmp/targets/e4dc114275694612ee236b231990d606b7879d05f64809611545c8234efb6cd4.doi-signing-key.pem
   ```
