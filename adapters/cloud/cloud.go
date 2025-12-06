// Package cloud provides a unified interface for cloud provider operations,
// abstracting away the specific implementations of AWS and OCI.
package cloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/abhissng/neuron/adapters/aws"
	"github.com/abhissng/neuron/adapters/oci"
)

// Provider represents the cloud provider type.
type Provider string

const (
	ProviderAWS Provider = "AWS"
	ProviderOCI Provider = "OCI"
)

// ErrUnsupportedProvider is returned when an invalid cloud provider is specified.
var ErrUnsupportedProvider = errors.New("cloud: unsupported provider")

// ErrNotInitialized is returned when the underlying manager is not initialized.
var ErrNotInitialized = errors.New("cloud: manager not initialized")

// CloudManager defines the common interface for cloud operations.
// It abstracts the underlying cloud provider (AWS or OCI).
type CloudManager interface {
	// Provider returns the cloud provider type.
	Provider() Provider

	// Object Storage Operations

	// UploadFile uploads data to cloud storage from a byte slice.
	// For AWS: bucket is the S3 bucket name, key is the object key.
	// For OCI: bucket is the bucket name, key is the object name (namespace is configured).
	UploadFile(ctx context.Context, bucket, key string, data []byte, contentType string, metadata map[string]string) error

	// UploadFileFromReader uploads data to cloud storage from an io.Reader.
	// This is ideal for streaming large files or multipart uploads.
	// contentLength should be provided if known; pass -1 if unknown (may incur buffering).
	UploadFileFromReader(ctx context.Context, bucket, key string, reader io.Reader, contentLength int64, contentType string, metadata map[string]string) error

	// DownloadFile downloads data from cloud storage.
	DownloadFile(ctx context.Context, bucket, key string) ([]byte, error)

	// ListObjects lists objects in a bucket with an optional prefix.
	ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)

	// DeleteObject deletes an object from cloud storage.
	DeleteObject(ctx context.Context, bucket, key string) error

	// GetPresignedURL generates a presigned URL for downloading an object.
	// Note: OCI implementation may differ; check provider-specific behavior.
	GetPresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error)

	// GetPresignedUploadURL generates a presigned URL for uploading an object.
	GetPresignedUploadURL(ctx context.Context, bucket, key, contentType string, expiration time.Duration) (string, error)

	// Secret Management Operations

	// GetSecret retrieves a secret value by its identifier.
	GetSecret(ctx context.Context, secretID string) (string, error)

	// CreateSecret creates a new secret.
	CreateSecret(ctx context.Context, name, value string) (string, error)

	// UpdateSecret updates an existing secret.
	UpdateSecret(ctx context.Context, secretID, value string) error

	// DeleteSecret deletes a secret.
	DeleteSecret(ctx context.Context, secretID string) error

	// GetMetadata returns metadata about the cloud manager configuration.
	GetMetadata() Metadata
}

// ObjectInfo represents metadata about a cloud storage object.
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
}

// Metadata contains information about the cloud manager configuration.
type Metadata struct {
	Provider Provider
	Region   string
}

// Config holds the configuration for creating a CloudManager.
type Config struct {
	// Provider specifies which cloud provider to use (AWS or OCI).
	Provider Provider

	// AWS-specific configuration (required if Provider is AWS).
	AWSConfig *aws.AWSConfig

	// AWS-Specific configuration options
	AWSOptions []aws.Option

	// OCI-specific configuration options (required if Provider is OCI).
	OCIOptions []oci.Option

	// OCINamespace is required for OCI object storage operations.
	OCINamespace string
}

func NewConfig() *Config {
	return &Config{}
}

// cloudManager is the concrete implementation of CloudManager.
type cloudManager struct {
	provider     Provider
	awsManager   *aws.AWSManager
	ociManager   *oci.OCIManager
	ociNamespace string
}

// NewCloudManager creates a new CloudManager based on the provided configuration.
// It returns an error if the provider is unsupported or if initialization fails.
func NewCloudManager(cfg Config) (CloudManager, error) {
	switch cfg.Provider {
	case ProviderAWS:
		return newAWSCloudManager(cfg)
	case ProviderOCI:
		return newOCICloudManager(cfg)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, cfg.Provider)
	}
}

// newAWSCloudManager initializes a CloudManager backed by AWS.
func newAWSCloudManager(cfg Config) (*cloudManager, error) {
	if cfg.AWSConfig == nil {
		return nil, errors.New("cloud: AWSConfig is required for AWS provider")
	}

	awsMgr, err := aws.NewAWSManager(*cfg.AWSConfig)
	if err != nil {
		return nil, fmt.Errorf("cloud: failed to initialize AWS manager: %w", err)
	}

	return &cloudManager{
		provider:   ProviderAWS,
		awsManager: awsMgr,
	}, nil
}

// newOCICloudManager initializes a CloudManager backed by OCI.
func newOCICloudManager(cfg Config) (*cloudManager, error) {
	if len(cfg.OCIOptions) == 0 {
		return nil, errors.New("cloud: OCIOptions are required for OCI provider")
	}

	ociMgr, err := oci.NewOCIManager(cfg.OCIOptions...)
	if err != nil {
		return nil, fmt.Errorf("cloud: failed to initialize OCI manager: %w", err)
	}

	return &cloudManager{
		provider:     ProviderOCI,
		ociManager:   ociMgr,
		ociNamespace: cfg.OCINamespace,
	}, nil
}

// NewCloudManagerFromExisting creates a CloudManager from existing manager instances.
// This is useful when you already have initialized AWS or OCI managers.
func NewCloudManagerFromExisting(awsMgr *aws.AWSManager, ociMgr *oci.OCIManager, ociNamespace string) (CloudManager, error) {
	if awsMgr != nil {
		return &cloudManager{
			provider:   ProviderAWS,
			awsManager: awsMgr,
		}, nil
	}

	if ociMgr != nil {
		return &cloudManager{
			provider:     ProviderOCI,
			ociManager:   ociMgr,
			ociNamespace: ociNamespace,
		}, nil
	}

	return nil, errors.New("cloud: at least one manager (AWS or OCI) must be provided")
}

// Provider returns the cloud provider type.
func (cm *cloudManager) Provider() Provider {
	return cm.provider
}

// GetMetadata returns metadata about the cloud manager configuration.
func (cm *cloudManager) GetMetadata() Metadata {
	meta := Metadata{
		Provider: cm.provider,
	}

	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager != nil {
			meta.Region = cm.awsManager.GetConfig().Region
		}
	case ProviderOCI:
		// OCI region would need to be exposed from OCIManager if needed
		meta.Region = ""
	}

	return meta
}
