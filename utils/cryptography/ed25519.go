package cryptography

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/abhissng/neuron/utils/helpers"
)

const (
	PRIVATE_KEY = "PRIVATE KEY"
	PUBLIC_KEY  = "PUBLIC KEY"
)

// GenerateEd25519KeyPair generates a new Ed25519 key pair and returns the PEM-encoded public and private keys.
func GenerateEd25519KeyPair() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate Ed25519 key pair: %v", err)
	}

	// Marshal private key to PKCS#8
	privateBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  PRIVATE_KEY,
		Bytes: privateBytes,
	})

	// Marshal public key to PKIX
	publicBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  PUBLIC_KEY,
		Bytes: publicBytes,
	})

	return string(publicKeyPEM), string(privateKeyPEM), nil
}

// LoadEd25519PrivateKey loads an Ed25519 private key from a PEM file or []byte content
func LoadEd25519PrivateKey(filePath string, content []byte) (ed25519.PrivateKey, error) {
	var data []byte
	var err error

	if !helpers.IsEmpty(filePath) {
		data, err = os.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %v", err)
		}
	} else {
		data = content
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	if block.Type != PRIVATE_KEY {
		return nil, fmt.Errorf("invalid PEM block type for private key: expected %q, got %q", PRIVATE_KEY, block.Type)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	privateKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("not a valid Ed25519 private key")
	}

	return privateKey, nil
}

// LoadEd25519PublicKey loads an Ed25519 public key from a PEM file or []byte content
func LoadEd25519PublicKey(filePath string, content []byte) (ed25519.PublicKey, error) {
	var data []byte
	var err error

	if !helpers.IsEmpty(filePath) {
		data, err = os.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return nil, fmt.Errorf("failed to read public key file: %v", err)
		}
	} else {
		data = content
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	if block.Type != PUBLIC_KEY {
		return nil, fmt.Errorf("invalid PEM block type for public key: expected %q, got %q", PUBLIC_KEY, block.Type)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	publicKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("not a valid Ed25519 public key")
	}

	return publicKey, nil
}

// GenerateAndSaveEd25519KeyPair generates a new Ed25519 key pair and saves
// the PEM-encoded keys to the specified files.
func GenerateAndSaveEd25519KeyPair(publicKeyPath, privateKeyPath string) error {
	privateKey, publicKey, err := GenerateEd25519KeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}
	// Save the private key to a file with restricted permissions.
	// 0600 means only the owner can read and write the file.
	if err := os.WriteFile(privateKeyPath, []byte(privateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key to file: %w", err)
	}

	// Save the public key to a file with standard read permissions.
	// 0600 means only the owner can read and write the file.
	if err := os.WriteFile(publicKeyPath, []byte(publicKey), 0600); err != nil {
		return fmt.Errorf("failed to write public key to file: %w", err)
	}

	return nil
}
