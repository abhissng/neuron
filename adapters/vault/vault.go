package vault

import (
	"context"
	"os"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	infisical "github.com/infisical/go-sdk"
	"github.com/infisical/go-sdk/packages/models"
)

// Vault struct holds the configuration for the Vault client
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
