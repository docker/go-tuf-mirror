# syntax=docker/dockerfile:1

#   Copyright Docker go-tuf-mirror authors
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build
WORKDIR /app

ARG TARGETOS TARGETARCH TARGETVARIANT
ARG VERSION="dev"
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH VARIANT=$TARGETVARIANT

RUN --mount=type=bind,source=.,target=/app \
  --mount=type=cache,target=$GOPATH/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags "-X main.version=$VERSION" -o /bin/go-tuf-mirror

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/go-tuf-mirror /go-tuf-mirror
ENTRYPOINT ["/go-tuf-mirror"]
