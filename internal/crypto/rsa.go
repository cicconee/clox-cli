package crypto

import (
	"crypto/rand"
	"crypto/rsa"
)

// RSA handles the RSA (Rivest–Shamir–Adleman) enryption.
type RSA struct{}

// Encrypt encrypts the data using the public key. The encrypted data is
// returned as a []byte.
func (r *RSA) Encrypt(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

// Decrypt decrypts the data using the private key. The decrypted data is
// returned as a []byte.
func (r *RSA) Decrypt(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}
