FROM golang:1.24-alpine AS base

# Install git and necessary build tools
RUN apk add --no-cache git make build-base

# Set up Go environment
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/deps

# Create a minimal module to download neuron dependencies
RUN echo 'package main\n\nimport (\n\t_ "github.com/abhissng/core-structures"\n\t_ "github.com/abhissng/neuron"\n)\n\nfunc main() {}\n' > main.go

# Set up token-based authentication (will be passed at build time)
ARG GITHUB_TOKEN
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Download dependencies
RUN go mod download

# Create directories and ensure they exist for later use
RUN mkdir -p /go/pkg/mod/github.com/abhissng

# List the downloaded modules for verification
RUN ls -la /go/pkg/mod/github.com/abhissng/ 
RUN go list -m github.com/abhissng/neuron
RUN go list -m github.com/abhissng/core-structures