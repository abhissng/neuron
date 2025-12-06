package oci

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// ========================= CONFIG =========================

type AuthMode int

const (
	AuthUserCreds AuthMode = iota
	AuthInstancePrincipal
)

type Config struct {
	Mode           AuthMode
	TenancyOCID    string
	UserOCID       string
	Fingerprint    string
	PrivateKeyPath string
	Passphrase     string
	Region         string
	Timeout        time.Duration
}

// ========================= CLIENT MANAGER =========================

type OCIManager struct {
	provider common.ConfigurationProvider
	config   *Config

	objectClient   *objectstorage.ObjectStorageClient
	computeClient  *core.ComputeClient
	identityClient *identity.IdentityClient

	enableObject   bool
	enableCompute  bool
	enableIdentity bool

	logger  *log.Log
	retries int
}

// ========================= OPTIONS =========================

type Option func(*OCIManager) error

func WithUserCredentials(tenancy, user, region, fingerprint, keyPath, passphrase string) Option {
	return func(cm *OCIManager) error {
		cm.config = &Config{
			Mode:           AuthUserCreds,
			TenancyOCID:    tenancy,
			UserOCID:       user,
			Fingerprint:    fingerprint,
			PrivateKeyPath: keyPath,
			Passphrase:     passphrase,
			Region:         region,
			Timeout:        2 * time.Minute,
		}
		return nil
	}
}

func WithInstancePrincipal(region string) Option {
	return func(cm *OCIManager) error {
		cm.config = &Config{
			Mode:    AuthInstancePrincipal,
			Region:  region,
			Timeout: 2 * time.Minute,
		}
		return nil
	}
}

func WithObjectStorage() Option {
	return func(cm *OCIManager) error {
		cm.enableObject = true
		return nil
	}
}

func WithCompute() Option {
	return func(cm *OCIManager) error {
		cm.enableCompute = true
		return nil
	}
}

func WithIdentity() Option {
	return func(cm *OCIManager) error {
		cm.enableIdentity = true
		return nil
	}
}

func WithLogger(logger *log.Log) Option {
	return func(cm *OCIManager) error {
		cm.logger = logger
		return nil
	}
}

func WithRetries(n int) Option {
	return func(cm *OCIManager) error {
		cm.retries = n
		return nil
	}
}

// ========================= INITIALIZER =========================

func NewOCIManager(opts ...Option) (*OCIManager, error) {
	cm := &OCIManager{}

	for _, opt := range opts {
		if err := opt(cm); err != nil {
			return nil, err
		}
	}
	if cm.logger == nil {
		return nil, errors.New("no logger provided")
	}

	if cm.config == nil {
		return nil, errors.New("no authentication configuration provided")
	}

	var provider common.ConfigurationProvider
	var err error

	switch cm.config.Mode {
	case AuthUserCreds:
		provider = common.NewRawConfigurationProvider(
			cm.config.TenancyOCID,
			cm.config.UserOCID,
			cm.config.Region,
			cm.config.Fingerprint,
			cm.config.PrivateKeyPath,
			common.String(cm.config.Passphrase),
		)
	case AuthInstancePrincipal:
		provider, err = auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to create instance principal provider: %v", err)
		}
	}

	cm.provider = provider

	if cm.enableObject {
		objClient, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
		if err != nil {
			return nil, err
		}
		objClient.SetRegion(cm.config.Region)
		cm.objectClient = &objClient
	}

	if cm.enableCompute {
		computeClient, err := core.NewComputeClientWithConfigurationProvider(provider)
		if err != nil {
			return nil, err
		}
		computeClient.SetRegion(cm.config.Region)
		cm.computeClient = &computeClient
	}

	if cm.enableIdentity {
		idClient, err := identity.NewIdentityClientWithConfigurationProvider(provider)
		if err != nil {
			return nil, err
		}
		idClient.SetRegion(cm.config.Region)
		cm.identityClient = &idClient
	}

	return cm, nil
}

// ========================= RETRY HELPER =========================

func (cm *OCIManager) withRetry(ctx context.Context, op func() error) error {
	var err error
	for i := 0; i < cm.retries; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err = op(); err == nil {
			return nil
		}
		cm.logger.Error("retry failed", log.Int("attempt", i+1), log.Int("max_attempts", cm.retries), log.Err(err))
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return err
}

func WithCtxTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, d)
}

// ========================= OBJECT STORAGE METHODS =========================

// UploadObjectFromReader uploads data from an io.Reader to OCI Object Storage.
// This method supports in-memory uploads and large files.
func (cm *OCIManager) UploadObjectFromReader(ctx context.Context, namespace, bucket, objectName string, reader io.Reader, contentLength int64, metadata map[string]string) error {
	if cm.objectClient == nil {
		return errors.New("object storage client not initialized")
	}

	// Convert io.Reader to io.ReadCloser if necessary
	var readCloser io.ReadCloser
	if rc, ok := reader.(io.ReadCloser); ok {
		readCloser = rc
	} else {
		readCloser = io.NopCloser(reader)
	}

	req := objectstorage.PutObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucket,
		ObjectName:    &objectName,
		PutObjectBody: readCloser,
		ContentLength: &contentLength,
	}

	if metadata != nil {
		req.OpcMeta = metadata
	}

	return cm.withRetry(ctx, func() error { _, e := cm.objectClient.PutObject(ctx, req); return e })
}

// UploadObject uploads a file from disk to OCI Object Storage.
// For in-memory uploads, use UploadObjectFromReader instead.
func (cm *OCIManager) UploadObject(ctx context.Context, namespace, bucket, objectName, filePath string) error {
	if cm.objectClient == nil {
		return errors.New("object storage client not initialized")
	}
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	stat, _ := f.Stat()

	req := objectstorage.PutObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucket,
		ObjectName:    &objectName,
		PutObjectBody: f,
		ContentLength: common.Int64(stat.Size()),
	}
	return cm.withRetry(ctx, func() error { _, e := cm.objectClient.PutObject(ctx, req); return e })
}

// DownloadObjectToMemory downloads an object from OCI Object Storage to memory.
// Returns the object content as a byte slice.
// Warning: For large objects, consider using DownloadObject to stream to disk instead.
func (cm *OCIManager) DownloadObjectToMemory(ctx context.Context, namespace, bucket, objectName string) ([]byte, error) {
	if cm.objectClient == nil {
		return nil, errors.New("object storage client not initialized")
	}

	resp, err := cm.objectClient.GetObject(ctx, objectstorage.GetObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucket,
		ObjectName:    &objectName,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Content.Close()
	}()

	data, err := io.ReadAll(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content: %w", err)
	}

	return data, nil
}

// DownloadObject downloads an object from OCI Object Storage to a file.
// For in-memory downloads, use DownloadObjectToMemory instead.
func (cm *OCIManager) DownloadObject(ctx context.Context, namespace, bucket, objectName, destPath string) error {
	if cm.objectClient == nil {
		return errors.New("object storage client not initialized")
	}
	resp, err := cm.objectClient.GetObject(ctx, objectstorage.GetObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucket,
		ObjectName:    &objectName,
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Content.Close()
	}()
	out, err := os.Create(filepath.Clean(destPath))
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()
	_, err = io.Copy(out, resp.Content)
	return err
}

func (cm *OCIManager) ListObjects(ctx context.Context, namespace, bucket string, prefix *string) ([]objectstorage.ObjectSummary, error) {
	if cm.objectClient == nil {
		return nil, errors.New("object storage client not initialized")
	}
	var result []objectstorage.ObjectSummary
	err := cm.withRetry(ctx, func() error {
		resp, e := cm.objectClient.ListObjects(ctx, objectstorage.ListObjectsRequest{
			NamespaceName: &namespace,
			BucketName:    &bucket,
			Prefix:        prefix,
		})
		if e != nil {
			return e
		}
		result = resp.Objects
		return nil
	})
	return result, err
}

func (cm *OCIManager) DeleteObject(ctx context.Context, namespace, bucket, objectName string) error {
	if cm.objectClient == nil {
		return errors.New("object storage client not initialized")
	}
	return cm.withRetry(ctx, func() error {
		_, e := cm.objectClient.DeleteObject(ctx, objectstorage.DeleteObjectRequest{
			NamespaceName: &namespace,
			BucketName:    &bucket,
			ObjectName:    &objectName,
		})
		return e
	})
}

func (cm *OCIManager) CreateBucket(ctx context.Context, namespace, compartmentOCID, bucketName, storageTier string) error {
	if cm.objectClient == nil {
		return errors.New("object storage client not initialized")
	}
	return cm.withRetry(ctx, func() error {
		_, e := cm.objectClient.CreateBucket(ctx, objectstorage.CreateBucketRequest{
			NamespaceName: &namespace,
			CreateBucketDetails: objectstorage.CreateBucketDetails{
				CompartmentId: &compartmentOCID,
				Name:          &bucketName,
				StorageTier:   objectstorage.CreateBucketDetailsStorageTierEnum(storageTier),
			},
		})
		return e
	})
}

func (cm *OCIManager) GetBucket(ctx context.Context, namespace, bucketName string) (*objectstorage.Bucket, error) {
	if cm.objectClient == nil {
		return nil, errors.New("object storage client not initialized")
	}
	var result *objectstorage.Bucket
	err := cm.withRetry(ctx, func() error {
		resp, e := cm.objectClient.GetBucket(ctx, objectstorage.GetBucketRequest{
			NamespaceName: &namespace,
			BucketName:    &bucketName,
		})
		if e != nil {
			return e
		}
		result = &resp.Bucket
		return nil
	})
	return result, err
}

func (cm *OCIManager) IsObjectExists(ctx context.Context, namespace, bucket, objectName string) (bool, error) {
	if cm.objectClient == nil {
		return false, errors.New("object storage client not initialized")
	}
	err := cm.withRetry(ctx, func() error {
		_, e := cm.objectClient.HeadObject(ctx, objectstorage.HeadObjectRequest{
			NamespaceName: &namespace,
			BucketName:    &bucket,
			ObjectName:    &objectName,
		})
		return e
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// ========================= COMPUTE METHODS =========================

func (cm *OCIManager) LaunchInstance(ctx context.Context, compartmentOCID, ad, shape, imageID, subnetID, displayName string) (*core.Instance, error) {
	if cm.computeClient == nil {
		return nil, errors.New("compute client not initialized")
	}
	var instance *core.Instance
	err := cm.withRetry(ctx, func() error {
		resp, e := cm.computeClient.LaunchInstance(ctx, core.LaunchInstanceRequest{
			LaunchInstanceDetails: core.LaunchInstanceDetails{
				CompartmentId:      &compartmentOCID,
				AvailabilityDomain: &ad,
				Shape:              &shape,
				ImageId:            &imageID,
				CreateVnicDetails:  &core.CreateVnicDetails{SubnetId: &subnetID},
				DisplayName:        &displayName,
			},
		})
		if e != nil {
			return e
		}
		instance = &resp.Instance
		return nil
	})
	return instance, err
}

func (cm *OCIManager) TerminateInstance(ctx context.Context, instanceID string) error {
	if cm.computeClient == nil {
		return errors.New("compute client not initialized")
	}
	return cm.withRetry(ctx, func() error {
		_, e := cm.computeClient.TerminateInstance(ctx, core.TerminateInstanceRequest{InstanceId: &instanceID})
		return e
	})
}

func (cm *OCIManager) ListInstances(ctx context.Context, compartmentOCID string) ([]core.Instance, error) {
	if cm.computeClient == nil {
		return nil, errors.New("compute client not initialized")
	}
	var instances []core.Instance
	err := cm.withRetry(ctx, func() error {
		resp, e := cm.computeClient.ListInstances(ctx, core.ListInstancesRequest{CompartmentId: &compartmentOCID})
		if e != nil {
			return e
		}
		instances = resp.Items
		return nil
	})
	return instances, err
}

// ========================= IDENTITY METHODS =========================

func (cm *OCIManager) ListCompartments(ctx context.Context, tenancyOCID string) ([]identity.Compartment, error) {
	if cm.identityClient == nil {
		return nil, errors.New("identity client not initialized")
	}
	var result []identity.Compartment
	err := cm.withRetry(ctx, func() error {
		resp, e := cm.identityClient.ListCompartments(ctx, identity.ListCompartmentsRequest{
			CompartmentId:          &tenancyOCID,
			AccessLevel:            identity.ListCompartmentsAccessLevelAccessible,
			CompartmentIdInSubtree: common.Bool(true),
		})
		if e != nil {
			return e
		}
		result = resp.Items
		return nil
	})
	return result, err
}
