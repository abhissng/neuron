package cloud_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/abhissng/neuron/adapters/aws"
	"github.com/abhissng/neuron/adapters/cloud"
	neuronlog "github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/adapters/oci"
)

// Example_awsCloudManager demonstrates creating a CloudManager with AWS.
func Example_awsCloudManager() {
	// Create AWS configuration
	awsCfg := &aws.AWSConfig{
		Region:          "us-east-1",
		AccessKeyID:     "your-access-key",
		SecretAccessKey: "your-secret-key",
	}

	// Create CloudManager with AWS provider
	manager, err := cloud.NewCloudManager(cloud.Config{
		Provider:  cloud.ProviderAWS,
		AWSConfig: awsCfg,
	})
	if err != nil {
		log.Fatalf("Failed to create cloud manager: %v", err)
	}

	// Use the unified interface
	ctx := context.Background()

	// Upload a file
	data := []byte("Hello, Cloud!")
	err = manager.UploadFile(ctx, "my-bucket", "hello.txt", data, "text/plain", nil)
	if err != nil {
		log.Printf("Upload failed: %v", err)
	}

	// Download a file
	downloaded, err := manager.DownloadFile(ctx, "my-bucket", "hello.txt")
	if err != nil {
		log.Printf("Download failed: %v", err)
	}
	fmt.Printf("Downloaded: %s\n", string(downloaded))

	// List objects
	objects, err := manager.ListObjects(ctx, "my-bucket", "")
	if err != nil {
		log.Printf("List failed: %v", err)
	}
	for _, obj := range objects {
		fmt.Printf("Object: %s (size: %d)\n", obj.Key, obj.Size)
	}

	// Generate presigned URL
	url, err := manager.GetPresignedURL(ctx, "my-bucket", "hello.txt", 15*time.Minute)
	if err != nil {
		log.Printf("Presign failed: %v", err)
	}
	fmt.Printf("Presigned URL: %s\n", url)

	// Get metadata
	meta := manager.GetMetadata()
	fmt.Printf("Provider: %s, Region: %s\n", meta.Provider, meta.Region)
}

// Example_ociCloudManager demonstrates creating a CloudManager with OCI.
func Example_ociCloudManager() {
	// Create a logger (required for OCI)
	logger := neuronlog.NewBasicLogger(false, true)

	// Create CloudManager with OCI provider
	manager, err := cloud.NewCloudManager(cloud.Config{
		Provider: cloud.ProviderOCI,
		OCIOptions: []oci.Option{
			oci.WithUserCredentials(
				"ocid1.tenancy.oc1..example",
				"ocid1.user.oc1..example",
				"us-ashburn-1",
				"aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99",
				"/path/to/private_key.pem",
				"",
			),
			oci.WithObjectStorage(),
			oci.WithLogger(logger),
			oci.WithRetries(3),
		},
		OCINamespace: "my-namespace",
	})
	if err != nil {
		log.Fatalf("Failed to create cloud manager: %v", err)
	}

	// Use the unified interface
	ctx := context.Background()

	// List objects (OCI)
	objects, err := manager.ListObjects(ctx, "my-bucket", "")
	if err != nil {
		log.Printf("List failed: %v", err)
	}
	for _, obj := range objects {
		fmt.Printf("Object: %s (size: %d)\n", obj.Key, obj.Size)
	}

	// Delete an object
	err = manager.DeleteObject(ctx, "my-bucket", "old-file.txt")
	if err != nil {
		log.Printf("Delete failed: %v", err)
	}
}

// Example_fromExistingManagers demonstrates creating a CloudManager from existing managers.
func Example_fromExistingManagers() {
	// Assume you already have an initialized AWS manager
	awsCfg := aws.AWSConfig{
		Region: "us-west-2",
	}
	existingAWSManager, err := aws.NewAWSManager(awsCfg)
	if err != nil {
		log.Fatalf("Failed to create AWS manager: %v", err)
	}

	// Wrap it in a CloudManager
	manager, err := cloud.NewCloudManagerFromExisting(existingAWSManager, nil, "")
	if err != nil {
		log.Fatalf("Failed to create cloud manager: %v", err)
	}

	// Now use the unified interface
	fmt.Printf("Provider: %s\n", manager.Provider())
}

// Example_secretManagement demonstrates secret management operations.
func Example_secretManagement() {
	awsCfg := &aws.AWSConfig{
		Region: "us-east-1",
	}

	manager, err := cloud.NewCloudManager(cloud.Config{
		Provider:  cloud.ProviderAWS,
		AWSConfig: awsCfg,
	})
	if err != nil {
		log.Fatalf("Failed to create cloud manager: %v", err)
	}

	ctx := context.Background()

	// Create a secret
	arn, err := manager.CreateSecret(ctx, "my-app/api-key", "super-secret-value")
	if err != nil {
		log.Printf("Create secret failed: %v", err)
	}
	fmt.Printf("Created secret ARN: %s\n", arn)

	// Get a secret
	value, err := manager.GetSecret(ctx, "my-app/api-key")
	if err != nil {
		log.Printf("Get secret failed: %v", err)
	}
	fmt.Printf("Secret value: %s\n", value)

	// Update a secret
	err = manager.UpdateSecret(ctx, "my-app/api-key", "new-secret-value")
	if err != nil {
		log.Printf("Update secret failed: %v", err)
	}

	// Delete a secret
	err = manager.DeleteSecret(ctx, "my-app/api-key")
	if err != nil {
		log.Printf("Delete secret failed: %v", err)
	}
}

// Example_streamingUpload demonstrates streaming large files using io.Reader.
func Example_streamingUpload() {
	awsCfg := &aws.AWSConfig{
		Region: "us-east-1",
	}

	manager, err := cloud.NewCloudManager(cloud.Config{
		Provider:  cloud.ProviderAWS,
		AWSConfig: awsCfg,
	})
	if err != nil {
		log.Fatalf("Failed to create cloud manager: %v", err)
	}

	ctx := context.Background()

	// Example 1: Upload from a bytes.Buffer (useful for in-memory data)
	buffer := bytes.NewBufferString("This is streaming data from a buffer!")
	err = manager.UploadFileFromReader(
		ctx,
		"my-bucket",
		"stream/buffer.txt",
		buffer,
		int64(buffer.Len()),
		"text/plain",
		map[string]string{"source": "buffer"},
	)
	if err != nil {
		log.Printf("Buffer upload failed: %v", err)
	}
	fmt.Println("Uploaded from buffer")

	// Example 2: Upload from an open file (useful for large files)
	file, err := os.Open("/path/to/large-file.dat")
	if err != nil {
		log.Printf("File open failed: %v", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("File close failed: %v", err)
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Failed to stat file: %v", err)
		return
	}
	err = manager.UploadFileFromReader(
		ctx,
		"my-bucket",
		"stream/large-file.dat",
		file,
		fileInfo.Size(),
		"application/octet-stream",
		nil,
	)
	if err != nil {
		log.Printf("File upload failed: %v", err)
	}
	fmt.Println("Uploaded large file")

	// Example 3: Upload with unknown size (AWS will buffer)
	dynamicData := bytes.NewReader([]byte("Dynamic content"))
	err = manager.UploadFileFromReader(
		ctx,
		"my-bucket",
		"stream/dynamic.txt",
		dynamicData,
		-1, // Unknown size - SDK will handle buffering
		"text/plain",
		nil,
	)
	if err != nil {
		log.Printf("Dynamic upload failed: %v", err)
	}
	fmt.Println("Uploaded dynamic content")
}

// Example_ociStreamingUpload demonstrates streaming uploads with OCI.
func Example_ociStreamingUpload() {
	logger := neuronlog.NewBasicLogger(false, true)

	manager, err := cloud.NewCloudManager(cloud.Config{
		Provider: cloud.ProviderOCI,
		OCIOptions: []oci.Option{
			oci.WithUserCredentials(
				"ocid1.tenancy.oc1..example",
				"ocid1.user.oc1..example",
				"us-ashburn-1",
				"aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99",
				"/path/to/private_key.pem",
				"",
			),
			oci.WithObjectStorage(),
			oci.WithLogger(logger),
			oci.WithRetries(3),
		},
		OCINamespace: "my-namespace",
	})
	if err != nil {
		log.Fatalf("Failed to create cloud manager: %v", err)
	}

	ctx := context.Background()

	// Upload from bytes.Buffer to OCI
	data := []byte("OCI streaming upload example")
	err = manager.UploadFileFromReader(
		ctx,
		"my-bucket",
		"oci-stream.txt",
		bytes.NewReader(data),
		int64(len(data)),
		"text/plain",
		map[string]string{"environment": "production"},
	)
	if err != nil {
		log.Printf("OCI upload failed: %v", err)
	}
	fmt.Println("Uploaded to OCI using streaming")

	// Download from OCI
	downloaded, err := manager.DownloadFile(ctx, "my-bucket", "oci-stream.txt")
	if err != nil {
		log.Printf("OCI download failed: %v", err)
	}
	fmt.Printf("Downloaded from OCI: %s\n", string(downloaded))
}
