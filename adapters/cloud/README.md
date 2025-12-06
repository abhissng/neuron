# Cloud Manager - Unified Cloud Provider Interface

The `cloud` package provides a unified interface for cloud operations, abstracting away the specific implementations of AWS and OCI.

## Features

- **Unified Interface**: Single `CloudManager` interface for both AWS and OCI
- **Streaming Support**: Upload and download using `io.Reader` for large files and multipart uploads
- **In-Memory Operations**: No need for file paths - work directly with bytes
- **Provider Abstraction**: Switch between AWS and OCI with minimal code changes
- **Error Handling**: Consistent error handling across providers

## Installation

```go
import "github.com/abhissng/neuron/adapters/cloud"
```

## Quick Start

### AWS Provider

```go
import (
    "context"
    "github.com/abhissng/neuron/adapters/aws"
    "github.com/abhissng/neuron/adapters/cloud"
)

// Create CloudManager with AWS
manager, err := cloud.NewCloudManager(cloud.Config{
    Provider: cloud.ProviderAWS,
    AWSConfig: &aws.AWSConfig{
        Region:          "us-east-1",
        AccessKeyID:     "your-key",
        SecretAccessKey: "your-secret",
    },
})

// Upload a file
data := []byte("Hello, Cloud!")
err = manager.UploadFile(ctx, "my-bucket", "hello.txt", data, "text/plain", nil)

// Download a file
downloaded, err := manager.DownloadFile(ctx, "my-bucket", "hello.txt")
```

### OCI Provider

```go
import (
    "github.com/abhissng/neuron/adapters/cloud"
    "github.com/abhissng/neuron/adapters/log"
    "github.com/abhissng/neuron/adapters/oci"
)

logger := log.NewBasicLogger(false, true)

manager, err := cloud.NewCloudManager(cloud.Config{
    Provider: cloud.ProviderOCI,
    OCIOptions: []oci.Option{
        oci.WithUserCredentials(
            "tenancy-ocid",
            "user-ocid",
            "us-ashburn-1",
            "fingerprint",
            "/path/to/key.pem",
            "",
        ),
        oci.WithObjectStorage(),
        oci.WithLogger(logger),
        oci.WithRetries(3),
    },
    OCINamespace: "my-namespace",
})
```

## Streaming Uploads

For large files or multipart uploads, use `UploadFileFromReader`:

### Upload from bytes.Buffer

```go
import "bytes"

buffer := bytes.NewBufferString("Streaming data!")
err := manager.UploadFileFromReader(
    ctx,
    "my-bucket",
    "stream.txt",
    buffer,
    int64(buffer.Len()),
    "text/plain",
    map[string]string{"source": "buffer"},
)
```

### Upload from File

```go
import "os"

file, err := os.Open("/path/to/large-file.dat")
defer file.Close()

fileInfo, _ := file.Stat()
err = manager.UploadFileFromReader(
    ctx,
    "my-bucket",
    "large-file.dat",
    file,
    fileInfo.Size(),
    "application/octet-stream",
    nil,
)
```

### Upload with Unknown Size

```go
// AWS will handle buffering automatically
reader := getDataReader() // any io.Reader
err := manager.UploadFileFromReader(
    ctx,
    "my-bucket",
    "dynamic.txt",
    reader,
    -1, // Unknown size
    "text/plain",
    nil,
)
```

## API Reference

### CloudManager Interface

```go
type CloudManager interface {
    // Provider returns the cloud provider type (AWS or OCI)
    Provider() Provider
    
    // Upload operations
    UploadFile(ctx context.Context, bucket, key string, data []byte, 
               contentType string, metadata map[string]string) error
    UploadFileFromReader(ctx context.Context, bucket, key string, 
                        reader io.Reader, contentLength int64, 
                        contentType string, metadata map[string]string) error
    
    // Download operation
    DownloadFile(ctx context.Context, bucket, key string) ([]byte, error)
    
    // Object management
    ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)
    DeleteObject(ctx context.Context, bucket, key string) error
    
    // Presigned URLs (AWS only)
    GetPresignedURL(ctx context.Context, bucket, key string, 
                    expiration time.Duration) (string, error)
    GetPresignedUploadURL(ctx context.Context, bucket, key, contentType string, 
                          expiration time.Duration) (string, error)
    
    // Secret management (AWS only)
    GetSecret(ctx context.Context, secretID string) (string, error)
    CreateSecret(ctx context.Context, name, value string) (string, error)
    UpdateSecret(ctx context.Context, secretID, value string) error
    DeleteSecret(ctx context.Context, secretID string) error
    
    // Metadata
    GetMetadata() Metadata
}
```

## Changes to Underlying Adapters

### AWS Adapter Changes

- **New Method**: `UploadToS3FromReader()` - Streams data from `io.Reader` to S3
- **Refactored**: `UploadToS3()` now uses `UploadToS3FromReader()` internally
- Supports streaming uploads with known or unknown content length

### OCI Adapter Changes

- **New Method**: `UploadObjectFromReader()` - Streams data from `io.Reader` to OCI Object Storage
- **New Method**: `DownloadObjectToMemory()` - Downloads object directly to memory as `[]byte`
- Existing file-based methods (`UploadObject`, `DownloadObject`) remain unchanged
- No more file path requirements for in-memory operations

## Benefits

### Memory Efficiency
- Stream large files without loading entire contents into memory
- Use `io.Reader` interface for efficient buffering

### Flexibility
- Work with `bytes.Buffer`, `os.File`, or any `io.Reader` implementation
- No temporary files needed for uploads

### Performance
- Optimal for large file transfers
- Reduced memory footprint with streaming

### Developer Experience
- Consistent API across AWS and OCI
- Easy to switch providers with minimal code changes
- Type-safe with compile-time guarantees

## Provider-Specific Notes

### AWS
- Supports all CloudManager methods
- Presigned URLs are fully supported
- Secret management via AWS Secrets Manager
- If content length is unknown (-1), AWS SDK will buffer the data

### OCI
- Object storage operations fully supported
- Presigned URLs: Not yet implemented
- Secret management: Not yet implemented (OCI Vault support needed)
- Content length is required for OCI uploads

## Migration Guide

### Before (Direct OCI Usage)
```go
// Required file path
err := ociManager.UploadObject(ctx, namespace, bucket, key, "/tmp/file.txt")
```

### After (Streaming with CloudManager)
```go
// Direct from memory
data := []byte("content")
err := cloudManager.UploadFile(ctx, bucket, key, data, "text/plain", nil)

// Or stream from reader
err := cloudManager.UploadFileFromReader(ctx, bucket, key, reader, size, "text/plain", nil)
```

## Error Handling

```go
manager, err := cloud.NewCloudManager(config)
if err != nil {
    if errors.Is(err, cloud.ErrUnsupportedProvider) {
        // Handle unsupported provider
    }
    if errors.Is(err, cloud.ErrNotInitialized) {
        // Handle uninitialized manager
    }
}
```

## Contributing

When adding new methods to the CloudManager interface:
1. Update the interface in `cloud.go`
2. Implement for AWS in `storage.go` or `secrets.go`
3. Implement for OCI (or return "not implemented" error)
4. Add examples in `example_test.go`
5. Update this README

## License

See the main neuron library LICENSE file.
