package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// GetS3Client returns the S3 client
func (a *AWSManager) GetS3Client() *s3.Client {
	return a.s3Client
}

// GetKMSClient returns the KMS client
func (a *AWSManager) GetKMSClient() *kms.Client {
	return a.kmsClient
}

// GetSecretsManagerClient returns the Secrets Manager client
func (a *AWSManager) GetSecretsManagerClient() *secretsmanager.Client {
	return a.secretsManager
}

// GetSSMClient returns the SSM client
func (a *AWSManager) GetSSMClient() *ssm.Client {
	return a.awsSSMClient
}

// GetConfig returns the AWS config
func (a *AWSManager) GetConfig() AWSConfig {
	return a.config
}
