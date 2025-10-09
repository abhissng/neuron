package vault

import (
	"strings"
	"time"

	"github.com/abhissng/neuron/adapters/aws"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/cryptography"
	"github.com/abhissng/neuron/utils/helpers"
)

// Option is a function that configures the Vault
type Option func(*Vault)

// WithAWSClient sets the AWS client
func WithAWSClient(client *aws.AWSManager) Option {
	return func(v *Vault) {
		v.awsClient = client
	}
}

// WithEnv sets the environment
func WithEnv(env string) Option {
	return func(v *Vault) {
		v.env = env
	}
}

// WithProjectID sets the project ID
func WithProjectID(projectID string) Option {
	return func(v *Vault) {
		v.projectID = projectID
	}
}

// WithPath sets the path
func WithPath(path string) Option {
	return func(v *Vault) {
		v.path = path
	}
}

// WithDefaultSource sets the default source and validates client requirements
func WithDefaultSource(source string) Option {
	return func(v *Vault) {
		source = strings.ToLower(source)
		if source != "aws" && source != "infisical" {
			helpers.Println(constant.WARN, "Invalid defaultSource: ", source, ", must be 'aws' or 'infisical'")
			source = "infisical"
		}
		v.defaultSource = source
	}
}

// WithTimeout sets the timeout duration
func WithTimeout(timeout time.Duration) Option {
	return func(v *Vault) {
		v.timeOut = timeout
	}
}

// WithCryptoManager sets the crypto manager
func WithCryptoManager(cryptoManager *cryptography.CryptoManager) Option {
	return func(v *Vault) {
		v.cryptoManager = cryptoManager
	}
}

// WithSiteURL sets the siteURL
func WithSiteURL(url string) Option {
	return func(v *Vault) {
		v.siteURL = url
	}
}
