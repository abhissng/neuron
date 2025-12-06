package cloud

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Default recovery window for secret deletion (in days).
const defaultSecretRecoveryWindow = 7 * 24 * time.Hour

// GetSecret retrieves a secret value by its identifier.
func (cm *cloudManager) GetSecret(ctx context.Context, secretID string) (string, error) {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return "", ErrNotInitialized
		}
		return cm.awsManager.GetSecret(ctx, secretID)

	case ProviderOCI:
		// TODO: Implement OCI Vault/Secrets service support in OCIManager.
		// OCI Vault/Secrets service is not implemented in the current OCIManager.
		// This would require adding OCI Vault client support.
		return "", errors.New("cloud: OCI secret management not implemented")

	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// CreateSecret creates a new secret.
func (cm *cloudManager) CreateSecret(ctx context.Context, name, value string) (string, error) {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return "", ErrNotInitialized
		}
		return cm.awsManager.CreateSecret(ctx, name, value)

	case ProviderOCI:
		// TODO: Implement OCI Vault/Secrets service support for CreateSecret in OCIManager.
		// OCI Vault/Secrets service is not implemented in the current OCIManager.
		return "", errors.New("cloud: OCI secret management not implemented")

	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// UpdateSecret updates an existing secret.
func (cm *cloudManager) UpdateSecret(ctx context.Context, secretID, value string) error {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return ErrNotInitialized
		}
		return cm.awsManager.UpdateSecret(ctx, secretID, value)

	case ProviderOCI:
		// TODO: Implement OCI Vault/Secrets service support for UpdateSecret in OCIManager.
		// OCI Vault/Secrets service is not implemented in the current OCIManager.
		return errors.New("cloud: OCI secret management not implemented")

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}

// DeleteSecret deletes a secret.
func (cm *cloudManager) DeleteSecret(ctx context.Context, secretID string) error {
	switch cm.provider {
	case ProviderAWS:
		if cm.awsManager == nil {
			return ErrNotInitialized
		}
		return cm.awsManager.DeleteSecret(ctx, secretID, defaultSecretRecoveryWindow)

	case ProviderOCI:
		// TODO: Implement OCI Vault/Secrets service support for DeleteSecret in OCIManager.
		// OCI Vault/Secrets service is not implemented in the current OCIManager.
		return errors.New("cloud: OCI secret management not implemented")

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedProvider, cm.provider)
	}
}
