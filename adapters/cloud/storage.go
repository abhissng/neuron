package cloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// UploadFile uploads data to cloud storage from a byte slice.
func (cm *cloudManager) UploadFile(ctx context.Context, bucket, key string, data []byte, contentType string, metadata map[string]string) error {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return ErrNotInitialized
		}
		_, err := cm.awsManager.UploadToS3(ctx, bucket, key, data, contentType, metadata)
		return err

	case ProviderOCI:
		if cm.ociManager == nil {
			return ErrNotInitialized
		}
		if cm.ociNamespace == "" {
			return errors.New("cloud: OCI namespace is required for object storage operations")
		}
		return cm.ociManager.UploadObjectFromReader(ctx, cm.ociNamespace, bucket, key, bytes.NewReader(data), int64(len(data)), metadata)

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// UploadFileFromReader uploads data to cloud storage from an io.Reader.
// This is ideal for streaming large files or multipart uploads.
func (cm *cloudManager) UploadFileFromReader(ctx context.Context, bucket, key string, reader io.Reader, contentLength int64, contentType string, metadata map[string]string) error {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return ErrNotInitialized
		}
		_, err := cm.awsManager.UploadToS3FromReader(ctx, bucket, key, reader, contentLength, contentType, metadata)
		return err

	case ProviderOCI:
		if cm.ociManager == nil {
			return ErrNotInitialized
		}
		if cm.ociNamespace == "" {
			return errors.New("cloud: OCI namespace is required for object storage operations")
		}
		return cm.ociManager.UploadObjectFromReader(ctx, cm.ociNamespace, bucket, key, reader, contentLength, metadata)

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// DownloadFile downloads data from cloud storage.
func (cm *cloudManager) DownloadFile(ctx context.Context, bucket, key string) ([]byte, error) {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return nil, ErrNotInitialized
		}
		return cm.awsManager.DownloadFromS3(ctx, bucket, key)

	case ProviderOCI:
		if cm.ociManager == nil {
			return nil, ErrNotInitialized
		}
		if cm.ociNamespace == "" {
			return nil, errors.New("cloud: OCI namespace is required for object storage operations")
		}
		return cm.ociManager.DownloadObjectToMemory(ctx, cm.ociNamespace, bucket, key)

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// ListObjects lists objects in a bucket with an optional prefix.
func (cm *cloudManager) ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return nil, ErrNotInitialized
		}
		objects, err := cm.awsManager.ListS3Objects(ctx, bucket, prefix)
		if err != nil {
			return nil, err
		}
		return convertAWSObjects(objects), nil

	case ProviderOCI:
		if cm.ociManager == nil {
			return nil, ErrNotInitialized
		}
		if cm.ociNamespace == "" {
			return nil, errors.New("cloud: OCI namespace is required for object storage operations")
		}
		var prefixPtr *string
		if prefix != "" {
			prefixPtr = &prefix
		}
		objects, err := cm.ociManager.ListObjects(ctx, cm.ociNamespace, bucket, prefixPtr)
		if err != nil {
			return nil, err
		}
		return convertOCIObjects(objects), nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// DeleteObject deletes an object from cloud storage.
func (cm *cloudManager) DeleteObject(ctx context.Context, bucket, key string) error {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return ErrNotInitialized
		}
		return cm.awsManager.DeleteS3Object(ctx, bucket, key)

	case ProviderOCI:
		if cm.ociManager == nil {
			return ErrNotInitialized
		}
		if cm.ociNamespace == "" {
			return errors.New("cloud: OCI namespace is required for object storage operations")
		}
		return cm.ociManager.DeleteObject(ctx, cm.ociNamespace, bucket, key)

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// GetPresignedURL generates a presigned URL for downloading an object.
func (cm *cloudManager) GetPresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return "", ErrNotInitialized
		}
		return cm.awsManager.CreateS3PresignedURL(ctx, bucket, key, expiration)

	case ProviderOCI:
		// TODO: Implement OCI presigned URL generation in the OCI adapter.
		// OCI presigned URLs require additional implementation in the OCI adapter.
		return "", errors.New("cloud: OCI presigned URL generation not implemented")

	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// GetPresignedUploadURL generates a presigned URL for uploading an object.
func (cm *cloudManager) GetPresignedUploadURL(ctx context.Context, bucket, key, contentType string, expiration time.Duration) (string, error) {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return "", ErrNotInitialized
		}
		return cm.awsManager.CreateS3PresignedPutURL(ctx, bucket, key, contentType, expiration)

	case ProviderOCI:
		// TODO: Implement OCI presigned upload URL generation in the OCI adapter.
		// OCI presigned URLs require additional implementation in the OCI adapter.
		return "", errors.New("cloud: OCI presigned upload URL generation not implemented")

	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// convertAWSObjects converts AWS S3 objects to the common ObjectInfo format.
func convertAWSObjects(objects []types.Object) []ObjectInfo {
	result := make([]ObjectInfo, len(objects))
	for i, obj := range objects {
		result[i] = ObjectInfo{
			Key:  safeString(obj.Key),
			Size: safeInt64(obj.Size),
			ETag: safeString(obj.ETag),
		}
		if obj.LastModified != nil {
			result[i].LastModified = *obj.LastModified
		}
	}
	return result
}

// convertOCIObjects converts OCI object summaries to the common ObjectInfo format.
func convertOCIObjects(objects []objectstorage.ObjectSummary) []ObjectInfo {
	result := make([]ObjectInfo, len(objects))
	for i, obj := range objects {
		result[i] = ObjectInfo{
			Key:  safeString(obj.Name),
			Size: safeInt64(obj.Size),
			ETag: safeString(obj.Etag),
		}
		if obj.TimeModified != nil {
			result[i].LastModified = obj.TimeModified.Time
		}
	}
	return result
}

// safeString safely dereferences a string pointer.
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeInt64 safely dereferences an int64 pointer.
func safeInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
