package vault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	neuron_aws "github.com/abhissng/neuron/adapters/aws"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	infisical "github.com/infisical/go-sdk"
)

// Constants for prefixes and timeouts
const (
	InfisicalPrefix      = "infisical:" // Optional prefix for Infisical (also default)
	SecretsManagerPrefix = "aws-sm:"
	ParameterStorePrefix = "aws-ssm:"
	AWSKMSPrefix         = "aws-kms:"
	timeout              = 30 * time.Second
)

// Vault struct holds configurations and clients for multiple secret backends.
type Vault struct {
	// Clients
	infisicalClient infisical.InfisicalClientInterface
	awsClient       *neuron_aws.AWSManager

	// Configuration
	env           string
	projectID     string
	path          string
	defaultSource string
	timeOut       time.Duration
}

// NewVault creates a new Vault with options
func NewVault(opts ...Option) *Vault {
	v := &Vault{
		// Set default values here
		timeOut:       timeout,
		defaultSource: "infisical", // Default to Infisical if not specified
	}

	// Apply all options
	for _, opt := range opts {
		opt(v)
	}

	if v.defaultSource == "aws" && v.awsClient == nil {
		helpers.Println(constant.ERROR, "awsClient is required when defaultSource is 'aws'")
		os.Exit(1)
	}

	if v.defaultSource == "infisical" {
		v.infisicalClient = infisical.NewInfisicalClient(context.Background(), infisical.Config{
			SiteUrl:          "https://app.infisical.com", // Optional, default is https://app.infisical.com
			AutoTokenRefresh: true,                        // Whether or not to let the SDK handle the access token lifecycle. Defaults to true if not specified.
		})
		// In case of blank client id and client secret
		// These values needs to be passed in environment variables with below key
		// INFISICAL_UNIVERSAL_AUTH_CLIENT_ID
		// INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET
		_, err := v.infisicalClient.Auth().UniversalAuthLogin("", "")
		if err != nil {
			helpers.Println(constant.ERROR, "Authentication failed with the infisical vault: ", err)
			os.Exit(1)
		}
	}

	// Check if at least one client is usable
	if v.infisicalClient == nil && v.awsClient.GetSecretsManagerClient() == nil && v.awsClient.GetSSMClient() == nil {
		helpers.Println(constant.ERROR, "no secret backend clients could be initialized")
		os.Exit(1)
	}

	return v
}

// === Backend Retrieval Functions ===

func (v *Vault) retrieveInfisicalSecret(key string) (string, error) {
	if v.infisicalClient == nil {
		return "", errors.New("infisical client not initialized")
	}
	secret, err := v.infisicalClient.Secrets().Retrieve(infisical.RetrieveSecretOptions{
		SecretKey:   key,
		Environment: v.env,
		ProjectID:   v.projectID,
		SecretPath:  v.path,
	})
	if err != nil {
		// Check if the error indicates "not found" - Infisical SDK might have specific errors
		// For now, assume any error here means failure, could refine later
		helpers.Println(constant.ERROR, "Error retrieving Infisical secret [", key, "]: ", err)
		// Consider checking for specific "not found" errors from the SDK if available
		// and returning ErrNotFound in that case.
		return "", fmt.Errorf("failed to retrieve Infisical secret %s: %w", key, err)
	}
	return secret.SecretValue, nil
}

func (v *Vault) retrieveAWSKMSSecret(ctx context.Context, secretId string) (string, error) {
	if v.awsClient.GetKMSClient() == nil {
		return "", errors.New("AWS KMS client not initialized")
	}

	awsCtx, cancel := context.WithTimeout(ctx, v.timeOut)
	defer cancel()

	result, err := v.awsClient.GetKMSKey(awsCtx, secretId)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret %s from AWS KMS: %w", secretId, err)
	}

	return result, nil
}

func (v *Vault) retrieveAWSSecretsManagerSecret(ctx context.Context, secretId string) (string, error) {
	if v.awsClient.GetSecretsManagerClient() == nil {
		return "", errors.New("AWS Secrets Manager client not initialized")
	}

	awsCtx, cancel := context.WithTimeout(ctx, v.timeOut)
	defer cancel()

	result, err := v.awsClient.GetSecret(awsCtx, secretId)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret %s from AWS Secrets Manager: %w", secretId, err)
	}

	return result, nil
}

func (v *Vault) retrieveAWSParameterStoreSecret(ctx context.Context, paramName string, withDecryption bool) (string, error) {
	if v.awsClient.GetSSMClient() == nil {
		return "", errors.New("AWS Parameter Store (SSM) client not initialized")
	}
	awsCtx, cancel := context.WithTimeout(ctx, v.timeOut)
	defer cancel()

	result, err := v.awsClient.GetParameter(awsCtx, paramName, withDecryption)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve parameter %s from AWS Parameter Store: %w", paramName, err)
	}

	return result, nil
}

// FetchVaultValue fetches a secret value from the configured backend based on key prefix.
// Prefixes: "aws-sm:", "aws-ssm:", "infisical:" (or no prefix defaults to Infisical).
func (v *Vault) FetchVaultValue(key string) (string, error) {
	var actualKey string
	// var source string
	ctx := context.Background()

	switch {
	case strings.HasPrefix(key, SecretsManagerPrefix):
		// source = "AWS Secrets Manager"
		actualKey = strings.TrimPrefix(key, SecretsManagerPrefix)
		// helpers.Println(constant.DEBUG, "Fetching from", source, " - Key:", actualKey)
		return v.retrieveAWSSecretsManagerSecret(ctx, actualKey)

	case strings.HasPrefix(key, ParameterStorePrefix):
		// source = "AWS Parameter Store"
		actualKey = strings.TrimPrefix(key, ParameterStorePrefix)
		// helpers.Println(constant.DEBUG, "Fetching from", source, " - Key:", actualKey)
		return v.retrieveAWSParameterStoreSecret(ctx, actualKey, true)

	case strings.HasPrefix(key, InfisicalPrefix):
		// source = "Infisical"
		actualKey = strings.TrimPrefix(key, InfisicalPrefix)
		// helpers.Println(constant.DEBUG, "Fetching from", source, "(explicit prefix) - Key:", actualKey)
		return v.retrieveInfisicalSecret(actualKey)
	case strings.HasPrefix(key, AWSKMSPrefix):
		// source = "AWS KMS"
		actualKey = strings.TrimPrefix(key, AWSKMSPrefix)
		// helpers.Println(constant.DEBUG, "Fetching from", source, "(explicit prefix) - Key:", actualKey)
		return v.retrieveAWSKMSSecret(ctx, actualKey)
	default:
		// Default to Infisical (or could be configured)
		// source = "Infisical"
		actualKey = key // Use the key as is
		if v.defaultSource == "aws" {
			return v.retrieveAWSParameterStoreSecret(ctx, actualKey, true)
		}
		// helpers.Println(constant.DEBUG, "Fetching from", source, "(default) - Key:", actualKey)
		return v.retrieveInfisicalSecret(actualKey)
	}
}

// OLD code for vault
// Vault struct holds the configuration for the Vault client
/*
type Vault struct {
	client    infisical.InfisicalClientInterface
	env       string
	projectId string
	path      string
}

// New Vault creates vault instance
func NewVault(env, projectId, path string) *Vault {
	if helpers.IsEmpty(path) {
		path = "/"
	}
	return &Vault{
		env:       env,
		projectId: projectId,
		path:      path,
	}
}

// Vault client initialization
func (v *Vault) InitVaultClient() {

	v.client = infisical.NewInfisicalClient(context.Background(), infisical.Config{
		SiteUrl:          "https://app.infisical.com", // Optional, default is https://app.infisical.com
		AutoTokenRefresh: true,                        // Whether or not to let the SDK handle the access token lifecycle. Defaults to true if not specified.
	})
	// In case of blank client id and client secret
	// These values needs to be passed in environment variables with below key
	// INFISICAL_UNIVERSAL_AUTH_CLIENT_ID
	// INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET
	_, err := v.client.Auth().UniversalAuthLogin("", "")
	if err != nil {
		helpers.Println(constant.ERROR, "Authentication failed with the vault: ", err)
		os.Exit(1)
	}

}

// retreiveSecret retrieves a secret from vault
func (v *Vault) retreiveSecret(key string) (models.Secret, error) {
	apiKeySecret, err := v.client.Secrets().Retrieve(infisical.RetrieveSecretOptions{
		SecretKey:   key,
		Environment: v.env,
		ProjectID:   v.projectId,
		SecretPath:  v.path,
	})
	if err != nil {
		helpers.Println(constant.ERROR, "Error retreiving secret ", key, " from vault: ", err)
		return models.Secret{}, err
	}

	return apiKeySecret, nil
}

// FetchVaultValue fetches a secret from vault
func (v *Vault) FetchVaultValue(key string) (string, error) {

	secret, err := v.retreiveSecret(key)
	if err != nil {
		helpers.Println(constant.ERROR, "Error fetching ", key, " values from vault: ", err)
		return "", err
	}

	return secret.SecretValue, nil
}

*/
