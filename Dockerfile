FROM --platform=linux/amd64 golang:1.22.0-alpine as deps

WORKDIR /

COPY go.* .

RUN go mod download;

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
