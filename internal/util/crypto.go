package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

const (
	EncryptionKeyEnv = "VI_MONGO_SECRET_KEY"

	DefaultEncryptionKey = "vi-mongo-default-encryption-key-please-change"

	KeyLength = 32 // AES-256
)

func GenerateEncryptionKey() (string, error) {
	key := make([]byte, KeyLength)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)
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
	key := os.Getenv(EncryptionKeyEnv)
	if key == "" {
		log.Warn().Msg("No encryption key found in environment variables, using default key")
		return DefaultEncryptionKey
	}
	return key
}

func EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	key := GetEncryptionKey()
	block, err := aes.NewCipher(createKey(key))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", nil
	}

	key := GetEncryptionKey()
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted password: %w", err)
	}

	block, err := aes.NewCipher(createKey(key))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}

	return string(plaintext), nil
}

func createKey(key string) []byte {
	k := make([]byte, 32)
	copy(k, []byte(key))
	return k
}
