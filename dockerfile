# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /neuron

# Copy module files first for efficient dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Optional: build your internal package (if needed)
# RUN go build -o /internal-app ./cmd/main.go

# Final image
FROM golang:1.24-alpine

WORKDIR /neuron

# Copy from builder
COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /neuron /neuron

# Set up environment variables
ENV GOPATH=/go
ENV GOMODCACHE=/go/pkg/mod

# Set entrypoint if needed
# ENTRYPOINT ["/internal-app"]