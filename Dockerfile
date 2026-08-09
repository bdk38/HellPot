# ==============================================================================
# Build Stage: Compile the Go source and prepare the filesystem
# ==============================================================================
FROM golang:1.25.8 AS build

WORKDIR /go/src/app

# Optional version injected by CI (e.g. v0.7.1). Falls back to "dev".
ARG VERSION=dev

# Cache dependencies
COPY go.* .
RUN go mod download

# Copy source
COPY . .

# Quality control
RUN go vet -v ./...
RUN go test -v ./...

# Build binary with versioning and optimization
RUN CGO_ENABLED=0 \
    go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /app \
    ./cmd/HellPot

# Prepare runtime directories and default config
RUN mkdir -p /config /logs && \
    cp docker_config.toml /config/config.toml

# ==============================================================================
# Final Stage: Minimal production image with a zero-privilege security profile
# ==============================================================================
FROM gcr.io/distroless/static-debian13

LABEL org.opencontainers.image.source="https://github.com/bdk38/HellPot"

COPY --from=build --chown=65532:65532 /app /app
COPY --from=build --chown=65532:65532 /config /config
COPY --from=build --chown=65532:65532 /logs /logs

VOLUME ["/config", "/logs"]
EXPOSE 8080
USER 65532

ENTRYPOINT ["/app", "-c", "/config/config.toml"]
