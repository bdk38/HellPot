# ==============================================================================
# Build Stage: Compile the Go source and prepares the filesystem
# ==============================================================================
#
# use newer golang because that's what I'm building with currently
FROM golang:1.25.7 AS build
#
# set the primary directory for build commands
WORKDIR /go/src/app
#
# cache dependencies
COPY go.* .
RUN go mod download
#
# copy source
COPY . .
#
# quality control
RUN go vet -v ./...
RUN go test -v ./...
#
# build binary with versioning and optimization outputting directly to /app for simplicity
RUN CGO_ENABLED=0 \
    VERSION=$(git tag --sort=-version:refname | head -n 1 || echo "dev") \
    go build -trimpath \
    -ldflags "-s -w -X main.version=$VERSION" \
    -o /app \
    cmd/HellPot/*.go
#
# prepare Workspace and create directories
RUN mkdir -p /config /logs && \
    cp docker_config.toml /config/config.toml
#
# ==============================================================================
# Final Stage: Minimal production image with a zero-privilege security profile
# ==============================================================================
#
# upgrade distro for lts
FROM gcr.io/distroless/static-debian13
#
# changed repo url
LABEL org.opencontainers.image.source="https://github.com/bdk38/HellPot"
#
# copy the binary and directories with non-root user
COPY --from=build --chown=65532:65532 /app /app
COPY --from=build --chown=65532:65532 /config /config
COPY --from=build --chown=65532:65532 /logs /logs
#
# add volumes for config and log files
VOLUME ["/config", "/logs"]
#
# expose default port
EXPOSE 8080
#
# run as non-root user
USER 65532
#
# execute the binary with the configuration flag as the default container process
ENTRYPOINT ["/app", "-c", "/config/config.toml"]
