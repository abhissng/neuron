# Cloud and File Adapters Usage

Packages covered:

- [`adapters/cloud`](../../adapters/cloud)
- [`adapters/aws`](../../adapters/aws)
- [`adapters/oci`](../../adapters/oci)
- [`adapters/file`](../../adapters/file)
- [`adapters/file/uploadFile`](../../adapters/file/uploadFile)

Primary files:

- [`adapters/cloud/cloud.go`](../../adapters/cloud/cloud.go)
- [`adapters/aws/aws.go`](../../adapters/aws/aws.go)
- [`adapters/oci/oci.go`](../../adapters/oci/oci.go)
- [`adapters/file/uploadFile/core.go`](../../adapters/file/uploadFile/core.go)

## Cloud Manager

Use `adapters/cloud` as unified interface over provider-specific adapters:

- choose provider (AWS/OCI)
- upload/download objects
- handle provider-agnostic error paths

```go
manager, err := cloud.NewCloudManager(cloud.Config{
	Provider: cloud.ProviderAWS,
	AWSConfig: &aws.AWSConfig{
		Region: "us-east-1",
	},
})
if err != nil {
	return err
}

err = manager.UploadFile(ctx, "bucket", "hello.txt", []byte("hello"), "text/plain", nil)
_ = err
```

Detailed provider examples are already documented in [`adapters/cloud/README.md`](../../adapters/cloud/README.md).

## AWS Adapter

`adapters/aws` wraps AWS integrations commonly used in services (storage, secrets, email-related components, and supporting client operations).

## OCI Adapter

`adapters/oci` wraps OCI object storage and related client options.

## File Adapter

`adapters/file` and `adapters/file/uploadFile` provide file processing and validation patterns:

- profile-based file validation
- MIME/extension/size constraints
- optional virus scanning integration

```go
validator, err := uploadFile.NewUploadFileValidator(
	uploadFile.WithProfile(uploadFile.ProfileImage),
	uploadFile.WithClamAV("localhost:3310"),
)
if err != nil {
	return err
}

err = validator.ValidateFile(fileHeader)
_ = err
```

See detailed usage in [`adapters/file/uploadFile/README.md`](../../adapters/file/uploadFile/README.md).

## Internal Helpers (implementation detail)

Provider wrappers include helper methods for request shaping, option defaults, and transport adaptation that are not stable external contracts.

## Caveats

- Do not perform blocking file scans on hot request paths without timeout controls.
- Keep cloud credentials and namespaces out of source code.
- For large payloads, prefer streaming APIs where available.
