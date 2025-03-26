FROM golang:1.24-alpine AS base

# Install git and necessary build tools
RUN apk add --no-cache git make build-base

# Set up Go environment
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/deps

# Set up token-based authentication (will be passed at build time)
ARG GITHUB_TOKEN
ARG NEURON_TAG
ARG CORE_TAG
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Create a minimal module to download neuron dependencies
RUN go env -w GOPRIVATE="github.com/abhissng*"

COPY go.mod go.sum ./

COPY . .

# Download dependencies explicitly
RUN go mod tidy -v \
    && go mod download

RUN cat > go.mod <<EOF
 module github.com/yourusername/neuron-deps/test
 
 go 1.24.1
 
 require (
    github.com/abhissng/neuron ${NEURON_TAG}
    github.com/abhissng/core-structures ${CORE_TAG}
)
EOF

# Download dependencies explicitly
RUN go mod tidy -v \
    && go mod download

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest 


# Verify that the dependencies were downloaded
RUN ls -la /go/pkg/mod/github.com/ || echo "Dependencies directory not found github.com"
RUN ls -la /go/pkg/mod/github.com/abhissng/ || echo "Dependencies directory not found github.com/abhissng"

RUN chmod -R 755 /go/pkg/mod 

# Create a new stage to ensure clean environment
FROM golang:1.24-alpine

# Copy the downloaded modules from the previous stage
COPY --from=base /go/pkg/mod/ /go/pkg/mod/

# # Create directory structure for validation
# RUN mkdir -p /go/pkg/mod/github.com/abhissng

# Set up Go environment in this stage too
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN chmod -R 755 /go/pkg/mod 

# Verify modules
RUN ls -la /go/pkg/mod/github.com/ || echo "Dependencies directory not copied properly for github.com"
RUN ls -la /go/pkg/mod/github.com/abhissng/ || echo "Dependencies directory not copied properly for github.com/abhissng"
