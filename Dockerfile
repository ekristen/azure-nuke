# syntax=docker/dockerfile:1.28-labs
ARG PROJECT_NAME=azure-nuke

FROM cgr.dev/chainguard/wolfi-base:latest AS base
ARG PROJECT_NAME
RUN apk add --no-cache ca-certificates
RUN addgroup -S ${PROJECT_NAME} && adduser -S ${PROJECT_NAME} -G ${PROJECT_NAME}

FROM docker.io/library/golang:1.27 AS build
ARG PROJECT_NAME
COPY / /src
WORKDIR /src
RUN \
  --mount=type=cache,target=/go/pkg \
  --mount=type=cache,target=/root/.cache/go-build \
  go build -o bin/${PROJECT_NAME} main.go

FROM base AS goreleaser
ARG PROJECT_NAME
COPY ${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}
USER ${PROJECT_NAME}
ENTRYPOINT ["/usr/local/bin/azure-nuke"]

FROM base
ARG PROJECT_NAME
COPY --from=build /src/bin/${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}
USER ${PROJECT_NAME}
ENTRYPOINT ["/usr/local/bin/azure-nuke"]