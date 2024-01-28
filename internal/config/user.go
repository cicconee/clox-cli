package config

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/cicconee/clox-cli/internal/security"
	"golang.org/x/crypto/bcrypt"
)

// User manages the user configuration values.
type User struct {
	passwordHash        string
	encryptedAPIToken   string
	encryptedPrivateKey string
	publicKey           string
}

// NewUser creates and returns a User. The public-private key pair will be generated
// for the user. The password is hashed. The api token and private key is encrypted.
//
// TODO:
//   - Encrypt the api token with the users password.
func NewUser(k *security.Keys, password string, apiToken string) (*User, error) {
	priv, pub, err := k.GenerateWithPassword(password)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hash(password)
	if err != nil {
		return nil, err
	}

	return &User{
		passwordHash:        string(hashedPassword),
		encryptedAPIToken:   apiToken,
		encryptedPrivateKey: string(priv),
		publicKey:           string(pub),
	}, nil
}

// VerifyPassword verifies if the password is correct. An error is returned if the password
// is incorrect. If correct it will return nil.
func (u *User) VerifyPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(password))
}

// RSAPrivateKey will decrypt this User's encrypted private key. It is returned as a
// *rsa.PrivateKey.
func (u *User) RSAPrivateKey(keys *security.Keys, password string) (*rsa.PrivateKey, error) {
	return keys.DecryptPrivateKey(u.encryptedPrivateKey, password)
}

// RSAPublicKey will decode this User's public key. It is returned as a *rsa.PublicKey.
func (u *User) RSAPublicKey(keys *security.Keys) (*rsa.PublicKey, error) {
	return keys.DecodePublicKey([]byte(u.publicKey))
}

// hash hashes the password.
func hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// UserConfigData is the structure used to marshal and unmarshal a User to JSON.
type UserConfigData struct {
	PasswordHash        string `json:"password"`
	EncryptedAPIToken   string `json:"api_token"`
	EncryptedPrivateKey string `json:"private_key"`
	PublicKey           string `json:"public_key"`
}

// UnmarshalJSON accepts a []byte which represents a users configuration and unmarshal
// it into this User.
func (u *User) UnmarshalJSON(data []byte) error {
	d := UserConfigData{}
	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}

	u.passwordHash = d.PasswordHash
	u.encryptedAPIToken = d.EncryptedAPIToken
	u.encryptedPrivateKey = d.EncryptedPrivateKey
	u.publicKey = d.PublicKey
	return nil
}

// MarshalJSON will marshal this user into JSON and return it as a []byte.
func (u *User) MarshalJSON() ([]byte, error) {
	d := UserConfigData{
		PasswordHash:        u.passwordHash,
		EncryptedAPIToken:   u.encryptedAPIToken,
		EncryptedPrivateKey: u.encryptedPrivateKey,
		PublicKey:           u.publicKey,
	}

	return json.Marshal(&d)
}
