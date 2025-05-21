package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

type EncryptionError struct {
	Operation string
	Err       error
}

func (e *EncryptionError) Error() string {
	return fmt.Sprintf("encryption error during %s: %v", e.Operation, e.Err)
}

func (e *EncryptionError) Unwrap() error {
	return e.Err
}

const (
	EncryptionKeyEnv = "VI_MONGO_SECRET_KEY"

	DefaultEncryptionKey = "vi-mongo-default-encryption-key-please-change"

	KeyLength = 32 // AES-256
)

func GenerateEncryptionKey() (string, error) {
	key := make([]byte, KeyLength)
	_, err := rand.Read(key)
	if err != nil {
		return "", &EncryptionError{Operation: "key generation", Err: err}
	}

	encodedKey := hex.EncodeToString(key)
	return encodedKey, nil
}

func PrintEncryptionKeyInstructions() {
	key, err := GenerateEncryptionKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate encryption key: %v\n", err)
		return
	}

	fmt.Println("Encryption key successfully generated for vi-mongo:")
	fmt.Println(key)
	fmt.Println("\nPlease store this key securely using one of the following methods:")

	fmt.Println("- Set it as an environment variable: VI_MONGO_SECRET_KEY")
	fmt.Println("- Save it to a file and reference the path in the config file")
	fmt.Println("  or use the CLI option: vi-mongo --secret-key=/path/to/key")
}

func GetEncryptionKey() string {
	return os.Getenv(EncryptionKeyEnv)
}

// EncryptPassword encrypts the given password using the provided hex-encoded key.
func EncryptPassword(password string, hexKey string) (string, error) {
	if password == "" {
		return "", nil
	}

	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", &EncryptionError{Operation: "key decoding", Err: err}
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", &EncryptionError{Operation: "cipher creation", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", &EncryptionError{Operation: "GCM creation", Err: err}
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", &EncryptionError{Operation: "nonce generation", Err: err}
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptPassword decrypts the hex-encoded encrypted password using the provided hex-encoded key.
func DecryptPassword(encryptedHex string, hexKey string) (string, error) {
	if encryptedHex == "" {
		return "", nil
	}

	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", &EncryptionError{Operation: "decode encrypted password", Err: err}
	}

	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", &EncryptionError{Operation: "key decoding", Err: err}
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", &EncryptionError{Operation: "cipher creation", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", &EncryptionError{Operation: "GCM creation", Err: err}
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", &EncryptionError{Operation: "ciphertext validation", Err: fmt.Errorf("ciphertext too short")}
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", &EncryptionError{Operation: "password decryption", Err: err}
	}

	return string(plaintext), nil
}
