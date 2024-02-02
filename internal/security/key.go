package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/cicconee/clox-cli/internal/crypto"
)

// Keys manages the RSA (Rivest–Shamir–Adleman) key pairs.
//
// The RSA key pairs are encrypted using the AES (Advanced Encryption Standard) with
// GCM (Galois/Counter Mode). This allows for the RSA private key to be encrypted and
// decrypted with a password.
type Keys struct {
	AES *crypto.AES
}

// GenerateWithPassword generates a password-encrypted RSA key pair. Only the private
// key is password protected. The first []byte returned is the private key, the second
// is the public key.
func (k *Keys) GenerateWithPassword(password string) ([]byte, []byte, error) {
	privKey, err := generateRSAKeyPair()
	if err != nil {
		return nil, nil, err
	}

	privKeyEncryted, err := k.encryptPrivateKey(privKey, password)
	if err != nil {
		return nil, nil, err
	}

	pubKeyPKCS := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)
	pubKeyPEM := encodePEM("RSA PUBLIC KEY", pubKeyPKCS)

	return privKeyEncryted, pubKeyPEM, nil
}

// encryptPrivateKey encrypts the private key with the password.
func (k *Keys) encryptPrivateKey(priv *rsa.PrivateKey, password string) ([]byte, error) {
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	encrypted, err := k.AES.EncryptWithPassword(privBytes, []byte(password))
	if err != nil {
		return nil, err
	}

	return encodePEM("RSA PRIVATE KEY", encrypted), nil
}

// DecryptPrivateKey decrypts the key with password and returns it as a *rsa.PrivateKey.
func (k *Keys) DecryptPrivateKey(encryptedKey, password string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(encryptedKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing encrypted key")
	}

	decrypted, err := k.AES.DecryptWithPassword(block.Bytes, []byte(password))
	if err != nil {
		return nil, err
	}

	privKey, err := x509.ParsePKCS1PrivateKey(decrypted)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

// DecodePublicKey will decode the key and return it s a *rsa.PublicKey.
func (k *Keys) DecodePublicKey(encodedKey []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(encodedKey)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing the key")
	}

	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// generateRSAKeyPair generates a RSA key pair.
func generateRSAKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// encodePEM encodes a pem.Block. The returned []byte is the ideal format for storing the data.
func encodePEM(t string, b []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  t,
		Bytes: b,
	})
}
