package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// Keys manages the RSA key pairs.
type Keys struct{}

// GenerateWithPassword generates a password-encrypted RSA key pair. Only the private
// key is password protected. The first []byte returned is the private key, the second
// is the public key.
func (k *Keys) GenerateWithPassword(password string) ([]byte, []byte, error) {
	privKey, err := generateRSAKeyPair()
	if err != nil {
		return nil, nil, err
	}

	privKeyEncryted, err := encryptPrivateKey(privKey, password)
	if err != nil {
		return nil, nil, err
	}

	pubKeyPKCS := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)
	pubKeyPEM := encodePEM("RSA PUBLIC KEY", pubKeyPKCS)

	return privKeyEncryted, pubKeyPEM, nil
}

// DecryptPrivateKey decrypts the key with password and returns it as a *rsa.PrivateKey.
func (k *Keys) DecryptPrivateKey(encryptedKey, password string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(encryptedKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing encrypted key")
	}

	// Assuming the first 16 bytes are the salt
	salt := block.Bytes[:16]
	encryptedData := block.Bytes[16:]

	key := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	decryptedData, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	privKey, err := x509.ParsePKCS1PrivateKey(decryptedData)
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

// encryptPrivateKey encrypts the private key with the password.
func encryptPrivateKey(priv *rsa.PrivateKey, password string) ([]byte, error) {
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nonce, nonce, privBytes, nil)

	return encodePEM("RSA PRIVATE KEY", append(salt, encrypted...)), nil
}

// encodePEM encodes a pem.Block. The returned []byte is the ideal format for storing the data.
func encodePEM(t string, b []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  t,
		Bytes: b,
	})
}
