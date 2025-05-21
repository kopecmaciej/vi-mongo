package util

import (
	"os"
	"testing"
)

func TestGenerateEncryptionKey(t *testing.T) {
	key, err := GenerateEncryptionKey()
	if err != nil {
		t.Fatalf("GenerateEncryptionKey failed: %v", err)
	}

	if len(key) != 2*KeyLength {
		t.Errorf("Expected key length %d, got %d", 2*KeyLength, len(key))
	}

	key2, err := GenerateEncryptionKey()
	if err != nil {
		t.Fatalf("Second GenerateEncryptionKey failed: %v", err)
	}

	if key == key2 {
		t.Error("Generated keys should be unique")
	}
}

func TestGetEncryptionKey(t *testing.T) {
	if err := os.Unsetenv(EncryptionKeyEnv); err != nil {
		t.Fatalf("Failed to unset environment variable: %v", err)
	}
	if key := GetEncryptionKey(); key != "" {
		t.Errorf("Expected empty key when env var not set, got %s", key)
	}

	testKey := "test-encryption-key"
	if err := os.Setenv(EncryptionKeyEnv, testKey); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	if key := GetEncryptionKey(); key != testKey {
		t.Errorf("Expected key %s, got %s", testKey, key)
	}
}

func TestEncryptDecryptPassword(t *testing.T) {
	key, err := GenerateEncryptionKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	testCases := []struct {
		name     string
		password string
		key      string
		wantErr  bool
	}{
		{
			name:     "Valid password and key",
			password: "my-secure-password",
			key:      key,
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			key:      key,
			wantErr:  false,
		},
		{
			name:     "Invalid key format",
			password: "password",
			key:      "not-a-hex-string",
			wantErr:  true,
		},
		{
			name:     "Key too short",
			password: "password",
			key:      "1234",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := EncryptPassword(tc.password, tc.key)
			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if _, ok := err.(*EncryptionError); !ok && err != nil {
					t.Errorf("Expected EncryptionError, got %T", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tc.password == "" && encrypted != "" {
				t.Error("Empty password should result in empty encrypted string")
			}

			if tc.password == "" {
				return
			}

			decrypted, err := DecryptPassword(encrypted, tc.key)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if decrypted != tc.password {
				t.Errorf("Expected decrypted password %s, got %s", tc.password, decrypted)
			}
		})
	}
}

func TestDecryptPassword_Errors(t *testing.T) {
	key, _ := GenerateEncryptionKey()

	testCases := []struct {
		name         string
		encryptedHex string
		key          string
		wantErr      bool
	}{
		{
			name:         "Empty encrypted password",
			encryptedHex: "",
			key:          key,
			wantErr:      false,
		},
		{
			name:         "Invalid hex in encrypted password",
			encryptedHex: "not-a-hex-string",
			key:          key,
			wantErr:      true,
		},
		{
			name:         "Ciphertext too short",
			encryptedHex: "1234",
			key:          key,
			wantErr:      true,
		},
		{
			name:         "Invalid key",
			encryptedHex: "1234567890abcdef",
			key:          "invalid-key",
			wantErr:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecryptPassword(tc.encryptedHex, tc.key)
			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if _, ok := err.(*EncryptionError); !ok && err != nil {
					t.Errorf("Expected EncryptionError, got %T", err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestEncryptionError(t *testing.T) {
	baseErr := os.ErrNotExist
	encErr := &EncryptionError{
		Operation: "test operation",
		Err:       baseErr,
	}

	expectedMsg := "encryption error during test operation: file does not exist"
	if encErr.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, encErr.Error())
	}

	unwrapped := encErr.Unwrap()
	if unwrapped != baseErr {
		t.Errorf("Expected unwrapped error to be %v, got %v", baseErr, unwrapped)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, err := GenerateEncryptionKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	passwords := []string{
		"simple",
		"Complex P@ssw0rd!",
		"Very long password with spaces and special characters: !@#$%^&*()",
		"Password with unicode: 你好世界",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			encrypted, err := EncryptPassword(password, key)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			decrypted, err := DecryptPassword(encrypted, key)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if decrypted != password {
				t.Errorf("Expected %q after round trip, got %q", password, decrypted)
			}

			encrypted2, _ := EncryptPassword(password, key)
			if encrypted == encrypted2 {
				t.Error("Encryption should not be deterministic")
			}
		})
	}
}
