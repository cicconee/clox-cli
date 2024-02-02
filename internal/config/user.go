package config

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/cicconee/clox-cli/internal/crypto"
	"github.com/cicconee/clox-cli/internal/security"
	"golang.org/x/crypto/bcrypt"
)

var ErrUnsetUser = errors.New("user not configured")

// User manages the user configuration values.
type User struct {
	passwordHash        string
	encryptedAPIToken   string
	encryptedPrivateKey string
	publicKey           string
	encryptedEncryptKey string
}

// NewUser creates and returns a User. The public-private key pair will be generated
// for the user. The password is hashed. The api token and private key is encrypted.
func NewUser(k *security.Keys, aes *crypto.AES, rsa *crypto.RSA, password string, apiToken string) (*User, error) {
	priv, pub, err := k.GenerateWithPassword(password)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hash(password)
	if err != nil {
		return nil, err
	}

	encryptedAPIToken, err := aes.EncryptWithPassword([]byte(apiToken), []byte(password))
	if err != nil {
		return nil, err
	}

	pubKey, err := k.DecodePublicKey(pub)
	if err != nil {
		return nil, err
	}

	encKey, err := aes.Generate()
	if err != nil {
		return nil, err
	}

	encryptedEncryptKey, err := rsa.Encrypt(encKey, pubKey)
	if err != nil {
		return nil, err
	}

	return &User{
		passwordHash:        string(hashedPassword),
		encryptedAPIToken:   base64.StdEncoding.EncodeToString(encryptedAPIToken),
		encryptedPrivateKey: string(priv),
		publicKey:           string(pub),
		encryptedEncryptKey: base64.StdEncoding.EncodeToString(encryptedEncryptKey),
	}, nil
}

// Validate validates that the user is completely configured. If any fields are not
// set it will return an error stating the field that is not set.
func (u *User) Validate() error {
	if u.passwordHash == "" {
		return errors.New("empty password")
	}

	if u.encryptedAPIToken == "" {
		return errors.New("empty api token")
	}

	if u.encryptedPrivateKey == "" {
		return errors.New("empty private key")
	}

	if u.publicKey == "" {
		return errors.New("empty public key")
	}

	return nil
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

// APIToken decrypts this User's encrypted API token.
func (u *User) APIToken(aes *crypto.AES, password string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(u.encryptedAPIToken)
	if err != nil {
		return "", err
	}

	token, err := aes.DecryptWithPassword(decoded, []byte(password))
	if err != nil {
		return "", err
	}

	return string(token), nil
}

func (u *User) EncryptKey(keys *security.Keys, rsa *crypto.RSA, password string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(u.encryptedEncryptKey)
	if err != nil {
		return nil, err
	}

	privKey, err := u.RSAPrivateKey(keys, password)
	if err != nil {
		return nil, err
	}

	return rsa.Decrypt(decoded, privKey)
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
	EncryptedEncryptKey string `json:"encrypt_key"`
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
	u.encryptedEncryptKey = d.EncryptedEncryptKey
	return nil
}

// MarshalJSON will marshal this user into JSON and return it as a []byte.
func (u *User) MarshalJSON() ([]byte, error) {
	d := UserConfigData{
		PasswordHash:        u.passwordHash,
		EncryptedAPIToken:   u.encryptedAPIToken,
		EncryptedPrivateKey: u.encryptedPrivateKey,
		PublicKey:           u.publicKey,
		EncryptedEncryptKey: u.encryptedEncryptKey,
	}

	return json.Marshal(&d)
}
