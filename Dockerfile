FROM --platform=linux/amd64 golang:1.22.0-alpine as deps

WORKDIR /

COPY go.* .

# This block can be replaced by `RUN go mod download` when github.com/docker/attest is public
RUN apk add --no-cache git
ENV GOPRIVATE="github.com/docker/attest"
RUN --mount=type=secret,id=GITHUB_TOKEN <<EOT
  set -e
  GITHUB_TOKEN=${GITHUB_TOKEN:-$(cat /run/secrets/GITHUB_TOKEN)}
  if [ -n "$GITHUB_TOKEN" ]; then
    echo "Setting GitHub access token"
    git config --global "url.https://x-access-token:${GITHUB_TOKEN}@github.com.insteadof" "https://github.com"
  fi
  go mod download
EOT

FROM --platform=$BUILDPLATFORM golang:1.22.0-alpine as build

WORKDIR /

COPY --from=deps --link $GOPATH/pkg/mod $GOPATH/pkg/mod
COPY . .

ARG TARGETOS TARGETARCH TARGETVARIANT
ARG VERSION="dev"
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH VARIANT=$TARGETVARIANT

RUN go build -ldflags "-X main.version=$VERSION" -o go-tuf-mirror;

FROM --platform=$TARGETPLATFORM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go-tuf-mirror /usr/local/bin/go-tuf-mirror
ENTRYPOINT ["/usr/local/bin/go-tuf-mirror"]
