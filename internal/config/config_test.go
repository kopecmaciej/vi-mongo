package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "vi-mongo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	configPath := filepath.Join(tmpDir, "test-config.yaml")

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return configPath, cleanup
}

func TestLoadConfigWithVersion_CustomPath(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := LoadConfigWithVersion("1.0.0", configPath)
	if err != nil {
		t.Fatalf("LoadConfigWithVersion failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.ConfigPath != configPath {
		t.Errorf("Expected ConfigPath %s, got %s", configPath, cfg.ConfigPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
}

func TestLoadConfigWithVersion_CustomPath_ExistingFile(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	customConfig := &Config{
		Version: "0.9.0",
		Log: LogConfig{
			Path:        "/custom/log/path.log",
			Level:       "debug",
			PrettyPrint: false,
		},
		ShowConnectionPage: false,
		ShowWelcomePage:    false,
	}

	data, err := yaml.Marshal(customConfig)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfigWithVersion("1.0.0", configPath)
	if err != nil {
		t.Fatalf("LoadConfigWithVersion failed: %v", err)
	}

	if cfg.Log.Path != "/custom/log/path.log" {
		t.Errorf("Expected log path /custom/log/path.log, got %s", cfg.Log.Path)
	}

	if cfg.Log.Level != "debug" {
		t.Errorf("Expected log level debug, got %s", cfg.Log.Level)
	}

	if cfg.ShowConnectionPage != false {
		t.Error("Expected ShowConnectionPage to be false")
	}

	if cfg.Version != "1.0.0" {
		t.Errorf("Expected version to be updated to 1.0.0, got %s", cfg.Version)
	}

	if cfg.ConfigPath != configPath {
		t.Errorf("Expected ConfigPath %s, got %s", configPath, cfg.ConfigPath)
	}
}

func TestUpdateConfig_CustomPath(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := LoadConfigWithVersion("1.0.0", configPath)
	if err != nil {
		t.Fatalf("LoadConfigWithVersion failed: %v", err)
	}

	cfg.ShowWelcomePage = false
	cfg.ShowConnectionPage = false
	cfg.Log.Level = "warn"

	if err := cfg.UpdateConfig(); err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var savedConfig Config
	if err := yaml.Unmarshal(data, &savedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if savedConfig.ShowWelcomePage != false {
		t.Error("Expected ShowWelcomePage to be false in saved config")
	}

	if savedConfig.ShowConnectionPage != false {
		t.Error("Expected ShowConnectionPage to be false in saved config")
	}

	if savedConfig.Log.Level != "warn" {
		t.Errorf("Expected log level warn in saved config, got %s", savedConfig.Log.Level)
	}
}

func TestAddConnection_CustomPath(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := LoadConfigWithVersion("1.0.0", configPath)
	if err != nil {
		t.Fatalf("LoadConfigWithVersion failed: %v", err)
	}

	mongoConfig := &MongoConfig{
		Name:     "test-connection",
		Host:     "localhost",
		Port:     27017,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
	}

	if err := cfg.AddConnection(mongoConfig); err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var savedConfig Config
	if err := yaml.Unmarshal(data, &savedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if len(savedConfig.Connections) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(savedConfig.Connections))
	}

	if savedConfig.Connections[0].Name != "test-connection" {
		t.Errorf("Expected connection name test-connection, got %s", savedConfig.Connections[0].Name)
	}

	if savedConfig.Connections[0].Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", savedConfig.Connections[0].Host)
	}
}

func TestSetCurrentConnection_CustomPath(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := LoadConfigWithVersion("1.0.0", configPath)
	if err != nil {
		t.Fatalf("LoadConfigWithVersion failed: %v", err)
	}

	mongoConfig := &MongoConfig{
		Name: "test-connection",
		Host: "localhost",
		Port: 27017,
	}

	if err := cfg.AddConnection(mongoConfig); err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}

	if err := cfg.SetCurrentConnection("test-connection"); err != nil {
		t.Fatalf("SetCurrentConnection failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var savedConfig Config
	if err := yaml.Unmarshal(data, &savedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if savedConfig.CurrentConnection != "test-connection" {
		t.Errorf("Expected current connection test-connection, got %s", savedConfig.CurrentConnection)
	}
}

func TestUpdateConnection_CustomPath(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := LoadConfigWithVersion("1.0.0", configPath)
	if err != nil {
		t.Fatalf("LoadConfigWithVersion failed: %v", err)
	}

	mongoConfig := &MongoConfig{
		Name: "test-connection",
		Host: "localhost",
		Port: 27017,
	}

	if err := cfg.AddConnection(mongoConfig); err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}

	updatedConfig := &MongoConfig{
		Name: "test-connection",
		Host: "remote-host",
		Port: 27018,
	}

	if err := cfg.UpdateConnection("test-connection", updatedConfig); err != nil {
		t.Fatalf("UpdateConnection failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var savedConfig Config
	if err := yaml.Unmarshal(data, &savedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if len(savedConfig.Connections) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(savedConfig.Connections))
	}

	if savedConfig.Connections[0].Host != "remote-host" {
		t.Errorf("Expected host remote-host, got %s", savedConfig.Connections[0].Host)
	}

	if savedConfig.Connections[0].Port != 27018 {
		t.Errorf("Expected port 27018, got %d", savedConfig.Connections[0].Port)
	}
}
