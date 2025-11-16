package util

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

type TestConfig struct {
	KeyField Key
	Nested   NestedConfig
}

type NestedConfig struct {
	KeyField Key
}

type Key struct {
	Keys        []string
	Runes       []string
	Description string
}

func TestMergeKeybindings(t *testing.T) {
	tests := []struct {
		name     string
		loaded   TestConfig
		default_ TestConfig
		expected TestConfig
	}{
		{
			name: "empty loaded config should use defaults",
			default_: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+C", "Ctrl+X"},
					Runes:       []string{"q"},
					Description: "Close application",
				},
			},
			loaded: TestConfig{},
			expected: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+C", "Ctrl+X"},
					Runes:       []string{"q"},
					Description: "Close application",
				},
			},
		},
		{
			name:     "empty default config should use loaded",
			default_: TestConfig{},
			loaded: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+V"},
					Runes:       []string{"q"},
					Description: "Paste",
				},
			},
			expected: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+V"},
					Runes:       []string{"q"},
					Description: "Paste",
				},
			},
		},
		{
			name: "loaded values with only keys should override defaults",
			default_: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+O"},
					Runes:       []string{"q"},
					Description: "Default action",
				},
			},
			loaded: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+N", "Ctrl+O"},
					Description: "Custom action",
				},
			},
			expected: TestConfig{
				KeyField: Key{
					Keys:        []string{"Ctrl+N", "Ctrl+O"},
					Runes:       nil,
					Description: "Custom action",
				},
			},
		},
		{
			name: "loaded values with only runes should override defaults",
			default_: TestConfig{
				KeyField: Key{
					Runes: []string{"p"},
				},
			},
			loaded: TestConfig{
				KeyField: Key{
					Runes: []string{"q", "w"},
				},
			},
			expected: TestConfig{
				KeyField: Key{
					Runes: []string{"q", "w"},
				},
			},
		},
		{
			name: "nested default values should be merged and overridden by loaded, missing fields should be set to defaults",
			default_: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        []string{"Ctrl+X"},
						Runes:       []string{"q"},
						Description: "Default nested",
					},
				},
				KeyField: Key{
					Runes: []string{"P"},
				},
			},
			loaded: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Runes: []string{"x"},
					},
				},
			},
			expected: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        nil,
						Runes:       []string{"x"},
						Description: "",
					},
				},
				KeyField: Key{
					Runes: []string{"P"},
				},
			},
		},
		{
			name: "nested loaded values should override defaults",
			default_: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        []string{"Ctrl+X"},
						Description: "Nuke my system",
					},
				},
			},
			loaded: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        []string{"Ctrl+Y"},
						Runes:       []string{"y"},
						Description: "Protect my system",
					},
				},
				KeyField: Key{
					Runes:       []string{"P"},
					Description: "I'm just a random key",
				},
			},
			expected: TestConfig{
				Nested: NestedConfig{
					KeyField: Key{
						Keys:        []string{"Ctrl+Y"},
						Runes:       []string{"y"},
						Description: "Protect my system",
					},
				},
				KeyField: Key{
					Runes:       []string{"P"},
					Description: "I'm just a random key",
				},
			},
		},
		{
			name:     "loaded values with empty key struct should be merged with defaults",
			default_: TestConfig{},
			loaded: TestConfig{
				KeyField: Key{},
			},
			expected: TestConfig{
				KeyField: Key{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaded := &tt.loaded
			MergeConfigs(loaded, &tt.default_)

			if !reflect.DeepEqual(*loaded, tt.expected) {
				t.Errorf("MergeConfigs() = %+v, want %+v", *loaded, tt.expected)
			}
		})
	}
}

func TestParseMongoDBURI(t *testing.T) {
	tests := []struct {
		name         string
		uri          string
		wantHost     string
		wantPort     string
		wantPassword string
		wantDb       string
		wantErr      bool
	}{
		{
			name:     "Valid standard URI",
			uri:      "mongodb://localhost:27017/mydatabase",
			wantHost: "localhost",
			wantPort: "27017",
			wantDb:   "mydatabase",
			wantErr:  false,
		},
		{
			name:         "Valid srv URI",
			uri:          "mongodb+srv://user:password@example.mongodb.net/mydatabase?retryWrites=true",
			wantHost:     "example.mongodb.net",
			wantPort:     "27017",
			wantPassword: "password",
			wantDb:       "mydatabase",
			wantErr:      false,
		},
		{
			name:     "Valid URI without port",
			uri:      "mongodb://localhost/mydatabase",
			wantHost: "localhost",
			wantPort: "27017",
			wantDb:   "mydatabase",
			wantErr:  false,
		},
		{
			name:     "Invalid prefix",
			uri:      "incorrect://localhost:27017/mydatabase",
			wantHost: "",
			wantPort: "",
			wantDb:   "",
			wantErr:  true,
		},
		{
			name:     "URI with options",
			uri:      "mongodb://localhost:27017/mydatabase?retryWrites=true",
			wantHost: "localhost",
			wantPort: "27017",
			wantDb:   "mydatabase",
			wantErr:  false,
		},
		{
			name:     "Sharded cluster URI with multiple hosts",
			uri:      "mongodb://mongodb1.example.com:27317,mongodb2.example.com:27017/mydatabase",
			wantHost: "mongodb1.example.com",
			wantPort: "27317",
			wantDb:   "mydatabase",
			wantErr:  false,
		},
		{
			name:         "Complex sharded cluster URI with options",
			uri:          "mongodb://myDatabaseUser:D1fficultP%40ssw0rd@mongodb1.example.com:27317,mongodb2.example.com:27017/?replicaSet=mySet&authSource=authDB&connectTimeoutMS=10000&authMechanism=SCRAM-SHA-1",
			wantHost:     "mongodb1.example.com",
			wantPort:     "27317",
			wantPassword: "D1fficultP%40ssw0rd",
			wantDb:       "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseMongoUri(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMongoDBURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if config.Host != tt.wantHost {
				t.Errorf("ParseMongoDBURI() gotHost = %v, want %v", config.Host, tt.wantHost)
			}
			if config.Port != tt.wantPort {
				t.Errorf("ParseMongoDBURI() gotPort = %v, want %v", config.Port, tt.wantPort)
			}
			if config.DB != tt.wantDb {
				t.Errorf("ParseMongoDBURI() gotDb = %v, want %v", config.DB, tt.wantDb)
			}
			if config.Password != tt.wantPassword {
				t.Errorf("ParseMongoDBURI()  = %v, want %v", config.DB, tt.wantDb)
			}
		})
	}
}

func TestValidateConfigPath(t *testing.T) {
	tmpDir := t.TempDir()

	tmpFile := tmpDir + "/test-config.yaml"
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty path is valid",
			path:    "",
			wantErr: false,
		},
		{
			name:    "Existing file is valid",
			path:    tmpFile,
			wantErr: false,
		},
		{
			name:    "Non-existent file in existing directory is valid",
			path:    tmpDir + "/new-config.yaml",
			wantErr: false,
		},
		{
			name:    "Non-existent file in non-existent directory is invalid",
			path:    tmpDir + "/nonexistent/config.yaml",
			wantErr: true,
			errMsg:  "config directory does not exist",
		},
		{
			name:    "Path pointing to directory is invalid",
			path:    tmpDir,
			wantErr: true,
			errMsg:  "config path is a directory",
		},
		{
			name:    "File in current directory is valid",
			path:    "config.yaml",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfigPath() error = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
