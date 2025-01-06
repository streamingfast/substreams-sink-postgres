FROM --platform=$BUILDPLATFORM golang:1.23-bullseye AS build

WORKDIR /src

ARG TARGETOS TARGETARCH VERSION=dev

RUN --mount=target=. \
      --mount=type=cache,target=/root/.cache/go-build \
      --mount=type=cache,target=/go/pkg \
      GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X \"main.version=$VERSION\"" -o /app/substreams-sink-sql ./cmd/substreams-sink-sql

FROM ubuntu:22.04

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
      apt-get -y install -y ca-certificates libssl3

COPY --from=build /app/substreams-sink-sql /app/substreams-sink-sql

ENTRYPOINT ["/app/substreams-sink-sql"]
