# syntax=docker/dockerfile:1.12-labs
FROM alpine:3.21.2 as base
ARG PROJECT_NAME=azure-nuke
RUN apk add --no-cache ca-certificates
RUN adduser -D azure-nuke

FROM ghcr.io/acorn-io/images-mirror/golang:1.21 AS build
COPY / /src
WORKDIR /src
RUN \
  --mount=type=cache,target=/go/pkg \
  --mount=type=cache,target=/root/.cache/go-build \
  go build -o bin/azure-nuke main.go

FROM base AS goreleaser
COPY azure-nuke /usr/local/bin/azure-nuke
USER azure-nuke

FROM base
COPY --from=build /src/bin/azure-nuke /usr/local/bin/azure-nuke
USER azure-nuke