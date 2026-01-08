########################################
# Base stage: download neuron deps
########################################
FROM golang:1.25.3-alpine3.22 AS base

# Install required tools
RUN apk add --no-cache git make build-base

# Docker-provided platform args
ARG TARGETOS
ARG TARGETARCH

# Go environment (DO NOT hardcode arch)
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH

WORKDIR /go/deps

# Build args
ARG GITHUB_TOKEN
ARG NEURON_TAG

# GitHub auth for private modules
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Private modules
RUN go env -w GOPRIVATE="github.com/abhissng*"

# Copy base module files
COPY go.mod go.sum ./
COPY . .

# Download dependencies explicitly
RUN go mod tidy -v \
    && go mod download

# Create minimal module only for neuron deps
RUN cat > go.mod <<EOF
module github.com/abhissng/neuron-deps/test

go 1.25.3

require (
    github.com/abhissng/neuron ${NEURON_TAG}
)
EOF

# Download neuron deps
RUN go mod tidy -v && go mod download

########################################
# Final image: only Go toolchain + deps
########################################
FROM golang:1.25.3-alpine3.22

ARG TARGETOS
ARG TARGETARCH

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH

# Copy cached modules
COPY --from=base /go/pkg/mod /go/pkg/mod

# Fix permissions (important for downstream builds)
RUN chmod -R 755 /go/pkg/mod/github.com/abhissng

# Optional verification (safe to keep)
RUN ls -la /go/pkg/mod/github.com/abhissng
