package cryptography

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const (
	RSA_PRIVATE_KEY = "RSA PRIVATE KEY"
)

// GenerateRSAKeypair generates an RSA private key of given bits (recommend 4096).
func GenerateRSAKeypair(bits int) (*rsa.PrivateKey, error) {
	if bits < 2048 {
		return nil, fmt.Errorf("bits too small: %d (minimum 2048 recommended)", bits)
	}
	return rsa.GenerateKey(rand.Reader, bits)
}

// SavePrivateKeyPEM saves a PKCS#1 PEM-encoded RSA private key.
func SavePrivateKeyPEM(path string, priv *rsa.PrivateKey) error {
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	block := &pem.Block{
		Type:  RSA_PRIVATE_KEY,
		Bytes: privBytes,
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, block)
}

// SavePublicKeyPEM saves a PKIX (X.509) PEM-encoded RSA public key.
func SavePublicKeyPEM(path string, pub *rsa.PublicKey) error {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return err
	}
	block := &pem.Block{
		Type:  PUBLIC_KEY,
		Bytes: pubBytes,
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, block)
}

// LoadRSAPrivateKeyPEM loads a PKCS#1 PEM-encoded RSA private key.
func LoadRSAPrivateKeyPEM(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != RSA_PRIVATE_KEY {
		return nil, errors.New("failed to decode PEM block containing RSA private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// LoadRSAPublicKeyPEM loads a PKIX PEM-encoded RSA public key.
func LoadRSAPublicKeyPEM(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != PUBLIC_KEY {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	pubIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := pubIfc.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return pub, nil
}

func GenerateAndSaveRSAKeypair(publicKeyPath, privateKeyPath string) error {
	priv, err := GenerateRSAKeypair(4096)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	fmt.Println("Saving keys to disk...")
	if err := SavePrivateKeyPEM(privateKeyPath, priv); err != nil {
		return fmt.Errorf("failed to save rsa private key: %w", err)
	}
	if err := SavePublicKeyPEM(publicKeyPath, &priv.PublicKey); err != nil {
		return fmt.Errorf("failed to save rsa public key: %w", err)
	}

	return nil
}
