package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// AES handles the AES (Advanced Encryption Standard) with GCM (Galois/Counter Mode)
// encryption.
type AES struct{}

// EncryptWithPassword encrypts data using the password. A unique salt is generated
// and used with the password to create the encryption key. The encrypted data is
// returned as a []byte. The salt is prepended to the encrypted data.
func (a *AES) EncryptWithPassword(data []byte, password []byte) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key(password, salt, 4096, 32, sha256.New)
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

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return append(salt, encrypted...), nil
}

// DecryptWithPassword decrypts data using the password. The salt and password
// are used to create the encryption key. The data is decrypted and returned as
// a []byte. Only the password used to encrypt the data will be able to decrypt it.
func (a *AES) DecryptWithPassword(data []byte, password []byte) ([]byte, error) {
	salt := data[:16]
	encryptedData := data[16:]

	key := pbkdf2.Key(password, salt, 4096, 32, sha256.New)
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
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Generates a random 32-byte key for AES encryption.
func (a *AES) Generate() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	return key, err
}
