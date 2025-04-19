package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmsTypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smTypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// ErrNotFound signifies that a secret/parameter was not found in the queried backend.
var ErrNotFound = errors.New("vault: secret/parameter not found")

// AWSConfig holds the configuration for AWS services
type AWSConfig struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	SessionToken     string
	Endpoint         string
	S3ForcePathStyle bool
}

func NewAwsConfig() *AWSConfig {
	return &AWSConfig{
		Region:           "ap-south-1",
		S3ForcePathStyle: false,
	}
}

// AWSManager provides a high-level interface to AWS services
type AWSManager struct {
	config         AWSConfig
	s3Client       *s3.Client
	kmsClient      *kms.Client
	secretsManager *secretsmanager.Client
	awsSSMClient   *ssm.Client
}

// Option is a function that configures the AWSManager
type Option func(*AWSManager)

// WithRegion sets the AWS region
func WithRegion(region string) Option {
	return func(w *AWSManager) {
		w.config.Region = region
	}
}

// WithEndpoint sets the custom endpoint for S3
func WithEndpoint(endpoint string) Option {
	return func(w *AWSManager) {
		w.config.Endpoint = endpoint
	}
}

// WithS3ForcePathStyle sets the S3 force path style
func WithS3ForcePathStyle(forcePathStyle bool) Option {
	return func(w *AWSManager) {
		w.config.S3ForcePathStyle = forcePathStyle
	}
}

// NewAWSWrapper creates a new instance of AWSManager with the provided options
func NewAWSWrapper(cfg AWSConfig, opts ...Option) (*AWSManager, error) {
	// Set default region if not provided
	if cfg.Region == "" {
		cfg.Region = "ap-south-1"
	}

	// Create AWS SDK configuration
	awsConfig, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create service clients
	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.S3ForcePathStyle
	})

	kmsClient := kms.NewFromConfig(awsConfig)
	secretsManagerClient := secretsmanager.NewFromConfig(awsConfig)
	ssmClient := ssm.NewFromConfig(awsConfig)

	// Apply options
	awsWrapper := &AWSManager{
		config:         cfg,
		s3Client:       s3Client,
		kmsClient:      kmsClient,
		secretsManager: secretsManagerClient,
		awsSSMClient:   ssmClient,
	}

	for _, opt := range opts {
		opt(awsWrapper)
	}

	return awsWrapper, nil
}

// loadAWSConfig creates the AWS SDK configuration
func loadAWSConfig(cfg AWSConfig) (aws.Config, error) {
	var awsConfig aws.Config
	var err error

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		// Use static credentials if provided
		awsConfig, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				cfg.SessionToken,
			)),
		)
	} else {
		// Otherwise, load from environment or AWS credential file
		awsConfig, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return aws.Config{}, err
	}

	return awsConfig, nil
}

// S3 Operations

// UploadToS3 uploads a file to an S3 bucket
func (a *AWSManager) UploadToS3(ctx context.Context, bucket, key string, data []byte, contentType string, metadata map[string]string) (*s3.PutObjectOutput, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	if metadata != nil {
		awsMetadata := make(map[string]string)
		for k, v := range metadata {
			awsMetadata[k] = v
		}
		input.Metadata = awsMetadata
	}

	result, err := a.s3Client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return result, nil
}

// DownloadFromS3 downloads a file from an S3 bucket
func (a *AWSManager) DownloadFromS3(ctx context.Context, bucket, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := a.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	return data, nil
}

// ListS3Objects lists objects in an S3 bucket
func (a *AWSManager) ListS3Objects(ctx context.Context, bucket, prefix string) ([]types.Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	result, err := a.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	return result.Contents, nil
}

// DeleteS3Object deletes an object from an S3 bucket
func (a *AWSManager) DeleteS3Object(ctx context.Context, bucket, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := a.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	return nil
}

// CreateS3PresignedURL creates a presigned URL for an S3 object
func (a *AWSManager) CreateS3PresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(a.s3Client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return request.URL, nil
}

// CreateS3PresignedPutURL creates a presigned URL for uploading to S3
func (a *AWSManager) CreateS3PresignedPutURL(ctx context.Context, bucket, key, contentType string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(a.s3Client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to create presigned put URL: %w", err)
	}

	return request.URL, nil
}

// KMS Operations

// EncryptWithKMS encrypts data using KMS
func (a *AWSManager) EncryptWithKMS(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	input := &kms.EncryptInput{
		KeyId:     aws.String(keyID),
		Plaintext: plaintext,
	}

	result, err := a.kmsClient.Encrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with KMS: %w", err)
	}

	return result.CiphertextBlob, nil
}

// DecryptWithKMS decrypts data using KMS
func (a *AWSManager) DecryptWithKMS(ctx context.Context, ciphertext []byte) ([]byte, error) {
	input := &kms.DecryptInput{
		CiphertextBlob: ciphertext,
	}

	result, err := a.kmsClient.Decrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with KMS: %w", err)
	}

	return result.Plaintext, nil
}

// CreateKMSKey creates a new KMS key
func (a *AWSManager) CreateKMSKey(ctx context.Context, description string) (string, error) {
	input := &kms.CreateKeyInput{
		Description: aws.String(description),
	}

	result, err := a.kmsClient.CreateKey(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create KMS key: %w", err)
	}

	return *result.KeyMetadata.KeyId, nil
}

// ListKMSKeys lists KMS keys
func (a *AWSManager) ListKMSKeys(ctx context.Context) ([]kmsTypes.KeyListEntry, error) {
	input := &kms.ListKeysInput{}

	result, err := a.kmsClient.ListKeys(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list KMS keys: %w", err)
	}

	return result.Keys, nil
}

// GetKMSKey retrieves a KMS key
func (a *AWSManager) GetKMSKey(ctx context.Context, keyID string) (string, error) {
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	}

	result, err := a.kmsClient.DescribeKey(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to describe KMS key %s: %w", keyID, err)
	}

	return *result.KeyMetadata.KeyId, nil
}

// Secrets Manager Operations

// GetSecret retrieves a secret from AWS Secrets Manager
func (a *AWSManager) GetSecret(ctx context.Context, secretID string) (string, error) {
	if a.secretsManager == nil {
		return "", errors.New("AWS Secrets Manager client not initialized")
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	result, err := a.secretsManager.GetSecretValue(ctx, input)
	if err != nil {
		var resourceNotFoundErr *smTypes.ResourceNotFoundException
		if errors.As(err, &resourceNotFoundErr) {
			helpers.Println(constant.DEBUG, "Secret not found in Secrets Manager:", secretID)
			return "", ErrNotFound
		}
		helpers.Println(constant.ERROR, "Failed to get secret [", secretID, "] from Secrets Manager: ", err)
		return "", fmt.Errorf("secrets manager get secret value failed for %s: %w", secretID, err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}
	if result.SecretBinary != nil {
		// Decide how to handle binary secrets. Base64 encode? Return bytes?
		// Returning base64 encoded string for consistency with string values.
		return base64.StdEncoding.EncodeToString(result.SecretBinary), nil
	}

	return *result.SecretString, nil
}

// CreateSecret creates a new secret in AWS Secrets Manager
func (a *AWSManager) CreateSecret(ctx context.Context, name, secretValue string) (string, error) {
	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(name),
		SecretString: aws.String(secretValue),
	}

	result, err := a.secretsManager.CreateSecret(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create secret: %w", err)
	}

	return *result.ARN, nil
}

// UpdateSecret updates an existing secret in AWS Secrets Manager
func (a *AWSManager) UpdateSecret(ctx context.Context, secretID, secretValue string) error {
	input := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretID),
		SecretString: aws.String(secretValue),
	}

	_, err := a.secretsManager.UpdateSecret(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// DeleteSecret deletes a secret from AWS Secrets Manager
func (a *AWSManager) DeleteSecret(ctx context.Context, secretID string, recoveryWindow time.Duration) error {
	input := &secretsmanager.DeleteSecretInput{
		SecretId:             aws.String(secretID),
		RecoveryWindowInDays: aws.Int64(int64(recoveryWindow.Hours() / 24)),
	}

	_, err := a.secretsManager.DeleteSecret(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// SSM Parameter Store Operations

// GetParameter retrieves a parameter from AWS SSM Parameter Store
func (a *AWSManager) GetParameter(ctx context.Context, name string, withDecryption bool) (string, error) {
	if a.awsSSMClient == nil {
		return "", errors.New("AWS Parameter Store (SSM) client not initialized")
	}
	input := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: &withDecryption,
	}

	result, err := a.awsSSMClient.GetParameter(ctx, input)
	if err != nil {
		var paramNotFoundErr *ssmTypes.ParameterNotFound
		if errors.As(err, &paramNotFoundErr) {
			helpers.Println(constant.DEBUG, "Parameter not found in Parameter Store:", name)
			return "", ErrNotFound
		}
		helpers.Println(constant.ERROR, "Failed to get parameter [", name, "] from Parameter Store: ", err)
		return "", fmt.Errorf("parameter store get parameter failed for %s: %w", name, err)
	}

	if result.Parameter == nil || result.Parameter.Value == nil {
		return "", errors.New("parameter value from Parameter Store is unexpectedly empty")
	}

	return *result.Parameter.Value, nil
}

// PutParameter creates or updates a parameter in AWS SSM Parameter Store
func (a *AWSManager) PutParameter(ctx context.Context, name, value, paramType string, overwrite bool) error {
	input := &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		Type:      ssmTypes.ParameterType(paramType),
		Overwrite: &overwrite,
	}

	_, err := a.awsSSMClient.PutParameter(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put parameter: %w", err)
	}

	return nil
}

// DeleteParameter deletes a parameter from AWS SSM Parameter Store
func (a *AWSManager) DeleteParameter(ctx context.Context, name string) error {
	input := &ssm.DeleteParameterInput{
		Name: aws.String(name),
	}

	_, err := a.awsSSMClient.DeleteParameter(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete parameter: %w", err)
	}

	return nil
}
