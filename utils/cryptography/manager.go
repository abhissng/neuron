// Package cryptography provides RSA-AES hybrid encryption utilities:
//   - GenerateRSAKeypair: create RSA private/public keypair (4096 bits)
//   - SavePrivateKeyPEM / SavePublicKeyPEM: write keys to disk as PEM
//   - LoadPrivateKeyPEM / LoadPublicKeyPEM: load keys from PEM files
//   - EncryptWithPublicKey: hybrid encrypt using RSA-OAEP (SHA-256) + AES-256-GCM
//   - DecryptWithPrivateKey: hybrid decrypt
//   - SignWithPrivateKey / VerifyWithPublicKey: optional signing using RSA-PSS (SHA-256)
//
// High-level format for encrypted blob returned by EncryptWithPublicKey:
//
//	[2 bytes big-endian len of encKey][encKey][12 bytes GCM nonce][ciphertext-with-gcm-tag]
//
// Security notes:
//   - RSA 4096 bits for envelope encryption (OAEP + SHA-256).
//   - AES-256-GCM for symmetric encryption (authenticated).
//   - Nonce is 12 bytes as required by GCM.
//   - Use crypto/rand for all randomness.
//   - Protect private keys (filesystem permissions, secret stores, etc.).
package cryptography

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

// CryptoManager wraps RSA key management, encryption, and signing.
type CryptoManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	hash       crypto.Hash
	keyBits    int
}

// Option defines a function type for functional options.
type Option func(*CryptoManager) error

// WithKeyPair allows providing existing RSA keys.
// If both private and public keys are provided, they will be used as is.
// If neither private nor public key is provided, we will need a environment variable to get the private and public key path
// MUST have the environment variable RSA_PRIVATE_KEY_PATH and RSA_PUBLIC_KEY_PATH set
func WithKeyPair(priv *rsa.PrivateKey, pub *rsa.PublicKey) Option {
	if priv == nil {
		rsaPrivateKey, err := LoadRSAPrivateKeyPEM(helpers.MustGetEnv(constant.RSA_PRIVATE_KEY_PATH))
		if err != nil {
			helpers.Printf(constant.FATAL, "Failed to load rsa private key %v", err)
		}
		priv = rsaPrivateKey
	}
	if pub == nil {
		rsaPublicKey, err := LoadRSAPublicKeyPEM(helpers.MustGetEnv(constant.RSA_PUBLIC_KEY_PATH))
		if err != nil {
			helpers.Printf(constant.FATAL, "Failed to load rsa public key %v", err)
		}
		pub = rsaPublicKey
	}
	return func(c *CryptoManager) error {
		c.privateKey = priv
		c.publicKey = pub
		return nil
	}
}

// WithPrivateKey allows providing existing RSA private key.
// MUST have the environment variable RSA_PRIVATE_KEY_PATH set if the priv is nil
func WithPrivateKey(priv *rsa.PrivateKey) Option {
	if priv == nil {
		rsaPrivateKey, err := LoadRSAPrivateKeyPEM(helpers.MustGetEnv(constant.RSA_PRIVATE_KEY_PATH))
		if err != nil {
			helpers.Printf(constant.FATAL, "Failed to load rsa private key %v", err)
		}
		priv = rsaPrivateKey
	}
	return func(c *CryptoManager) error {
		c.privateKey = priv
		return nil
	}
}

// WithPublicKey allows providing existing RSA public key.
// MUST have the environment variable RSA_PUBLIC_KEY_PATH set if the pub is nil
func WithPublicKey(pub *rsa.PublicKey) Option {
	if pub == nil {
		rsaPublicKey, err := LoadRSAPublicKeyPEM(helpers.MustGetEnv(constant.RSA_PUBLIC_KEY_PATH))
		if err != nil {
			helpers.Printf(constant.FATAL, "Failed to load rsa public key %v", err)
		}
		pub = rsaPublicKey
	}
	return func(c *CryptoManager) error {
		c.publicKey = pub
		return nil
	}
}

// WithKeySize sets RSA key size (default: 4096).
func WithKeySize(bits int) Option {
	return func(c *CryptoManager) error {
		if bits < 2048 {
			return fmt.Errorf("key size too small: %d", bits)
		}
		c.keyBits = bits
		return nil
	}
}

// WithHash sets the hash algorithm (default: SHA256).
func WithHash(hash crypto.Hash) Option {
	return func(c *CryptoManager) error {
		c.hash = hash
		return nil
	}
}

// NewCryptoManager initializes a CryptoManager with optional parameters.
func NewCryptoManager(opts ...Option) (*CryptoManager, error) {
	cm := &CryptoManager{
		hash:    crypto.SHA256,
		keyBits: 2048,
	}
	for _, opt := range opts {
		if err := opt(cm); err != nil {
			return nil, err
		}
	}
	if cm.privateKey == nil {
		priv, err := rsa.GenerateKey(rand.Reader, cm.keyBits)
		if err != nil {
			return nil, err
		}
		cm.privateKey = priv
		cm.publicKey = &priv.PublicKey
	}
	return cm, nil
}

// PublicKey returns the public key.
func (c *CryptoManager) PublicKey() *rsa.PublicKey {
	return c.publicKey
}

// PrivateKey returns the private key.
func (c *CryptoManager) PrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

// Encrypt performs hybrid encryption:
//   - Generates ephemeral AES-256 key
//   - Encrypts plaintext using AES-GCM
//   - Encrypts AES key using RSA-OAEP(SHA-256)
//   - Returns concatenation: [2-byte len][encKey][12-byte nonce][ciphertext+tag]
func (c *CryptoManager) Encrypt(plaintext []byte) (string, error) {
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	hash := sha256.New()
	encKey, err := rsa.EncryptOAEP(hash, rand.Reader, c.publicKey, aesKey, nil)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := binary.Write(&out, binary.BigEndian, uint16(len(encKey))); err != nil {
		return "", err
	}
	out.Write(encKey)
	out.Write(nonce)
	out.Write(ciphertext)
	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

// Decrypt reverses Encrypt:
//   - Reads encKey length + encKey
//   - Decrypts encKey using RSA-OAEP to obtain AES key
//   - Uses AES-GCM with nonce to decrypt ciphertext
func (c *CryptoManager) Decrypt(value string) ([]byte, error) {
	blob, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(blob)
	var encKeyLen uint16
	if err := binary.Read(reader, binary.BigEndian, &encKeyLen); err != nil {
		return nil, err
	}
	encKey := make([]byte, encKeyLen)
	if _, err := io.ReadFull(reader, encKey); err != nil {
		return nil, err
	}
	hash := sha256.New()
	aesKey, err := rsa.DecryptOAEP(hash, rand.Reader, c.privateKey, encKey, nil)
	if err != nil {
		return nil, err
	}

	remaining, _ := io.ReadAll(reader)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(remaining) < nonceSize {
		return nil, errors.New("invalid ciphertext")
	}
	nonce, ciphertext := remaining[:nonceSize], remaining[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Sign signs the given data using RSA-PSS.
func (c *CryptoManager) Sign(data []byte) ([]byte, error) {
	h := sha256.Sum256(data)
	return rsa.SignPSS(rand.Reader, c.privateKey, c.hash, h[:], nil)
}

// Verify verifies a signature using the public key.
func (c *CryptoManager) Verify(data, sig []byte) error {
	h := sha256.Sum256(data)
	return rsa.VerifyPSS(c.publicKey, c.hash, h[:], sig, nil)
}

// SaveKeys saves private and public keys to PEM files.
func (c *CryptoManager) SaveKeys(privPath, pubPath string) error {
	if err := SavePrivateKeyPEM(privPath, c.privateKey); err != nil {
		return err
	}
	return SavePublicKeyPEM(pubPath, c.publicKey)
}

/*
//// Note:
//// The SignWithPrivateKey / VerifyWithPublicKey functions above attempt to be illustrative.
//// In practice, just call rsa.SignPSS/VerifyPSS with crypto.SHA256 directly (see example below).

// ---------------------------
// Example usage in main()
// ---------------------------

func main() {
	// Example: generate keys, save to disk, encrypt and decrypt.
	const (
		privPath = "id_rsa_neuron.pem"
		pubPath  = "id_rsa_neuron.pub.pem"
	)

	fmt.Println("Generating RSA-4096 keypair...")
	priv, err := GenerateRSAKeypair(4096)
	if err != nil {
		panic(err)
	}

	fmt.Println("Saving keys to disk...")
	if err := SavePrivateKeyPEM(privPath, priv); err != nil {
		panic(err)
	}
	if err := SavePublicKeyPEM(pubPath, &priv.PublicKey); err != nil {
		panic(err)
	}

	fmt.Println("Loading public key...")
	pub, err := LoadPublicKeyPEM(pubPath)
	if err != nil {
		panic(err)
	}

	plain := []byte("Sensitive message: rotate keys and protect private key files!")
	fmt.Printf("Plaintext: %s\n", plain)

	fmt.Println("Encrypting with public key...")
	blob, err := EncryptWithPublicKey(pub, plain)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Encrypted blob size: %d bytes\n", len(blob))

	fmt.Println("Loading private key for decryption...")
	priv2, err := LoadPrivateKeyPEM(privPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("Decrypting with private key...")
	decrypted, err := DecryptWithPrivateKey(priv2, blob)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Decrypted plaintext: %s\n", decrypted)

	// Optional: demonstrate signing & verifying with rsa.SignPSS / rsa.VerifyPSS directly
	fmt.Println("Demonstrating sign & verify (RSA-PSS + SHA-256)...")
	data := []byte("message to sign")
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPSS(rand.Reader, priv2, crypto.SHA256, hashed[:], nil)
	if err != nil {
		panic(err)
	}
	// verify
	if err := rsa.VerifyPSS(&priv2.PublicKey, crypto.SHA256, hashed[:], signature, nil); err != nil {
		panic("signature verification failed: " + err.Error())
	}
	fmt.Println("Signature verified successfully.")
}
*/
