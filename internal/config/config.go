package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kopecmaciej/vi-mongo/internal/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	ConfigFile = "config.yaml"
	LogPath    = "/tmp/vi-mongo.log"
)

type MongoOptions struct {
	AlwaysConfirmActions  *bool  `yaml:"alwaysConfirmActions,omitempty"`
	AuthorizedDatabases   *bool  `yaml:"authorizedDatabases,omitempty"`
	AuthorizedCollections *bool  `yaml:"authorizedCollections,omitempty"`
	Limit                 *int64 `yaml:"limit,omitempty"`
}

type MongoConfig struct {
	Uri      string       `yaml:"url"`
	Host     string       `yaml:"host"`
	Port     int          `yaml:"port"`
	Database string       `yaml:"database"`
	Username string       `yaml:"username"`
	Password string       `yaml:"password"`
	Name     string       `yaml:"name"`
	Timeout  int          `yaml:"timeout"`
	Options  MongoOptions `yaml:"options"`
}

type LogConfig struct {
	Path        string `yaml:"path"`
	Level       string `yaml:"level"`
	PrettyPrint bool   `yaml:"prettyPrint"`
}

type EditorConfig struct {
	Command string `yaml:"command"`
	Env     string `yaml:"env"`
}

type StylesConfig struct {
	BetterSymbols bool   `yaml:"betterSymbols"`
	CurrentStyle  string `yaml:"currentStyle"`
}

type Config struct {
	Version            string        `yaml:"version"`
	Log                LogConfig     `yaml:"log"`
	Editor             EditorConfig  `yaml:"editor"`
	ShowConnectionPage bool          `yaml:"showConnectionPage"`
	ShowWelcomePage    bool          `yaml:"showWelcomePage"`
	CurrentConnection  string        `yaml:"currentConnection"`
	Connections        []MongoConfig `yaml:"connections"`
	Styles             StylesConfig  `yaml:"styles"`
}

// LoadConfig loads the config file
// If the file does not exist, it will be created
// with the default settings
func LoadConfig() (*Config, error) {
	defaultConfig := &Config{}
	defaultConfig.loadDefaults()

	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	return util.LoadConfigFile(defaultConfig, configPath)
}

// loadDefaults loads the default config settings
func (c *Config) loadDefaults() {
	c.Version = "1.0.0"
	c.Log = LogConfig{
		Path:        LogPath,
		Level:       "info",
		PrettyPrint: true,
	}
	c.Editor = EditorConfig{
		Command: "",
		Env:     "EDITOR",
	}
	c.Styles = StylesConfig{
		BetterSymbols: true,
		CurrentStyle:  "default.yaml",
	}
	c.ShowConnectionPage = true
	c.ShowWelcomePage = true
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	configPath, err := util.GetConfigDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", configPath, ConfigFile), nil
}

// UpdateConfig updates the config file with the new settings
func (c *Config) UpdateConfig() error {
	updatedConfig, err := yaml.Marshal(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal config")
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, updatedConfig, 0644); err != nil {
		log.Error().Err(err).Msg("Failed to write config file")
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetEditorCmd returns the editor command from the config file
func (c *Config) GetEditorCmd() (string, error) {
	if c.Editor.Env == "" && c.Editor.Command == "" {
		return "", fmt.Errorf("editor not set")
	}
	if c.Editor.Command != "" {
		return c.Editor.Command, nil
	}

	return os.Getenv(c.Editor.Env), nil
}

// SetCurrentConnection sets the current connection in the config file
func (c *Config) SetCurrentConnection(name string) error {
	// If the user has set the alwaysShowConnectionPage setting to true,
	// we don't want to save the current connection
	c.CurrentConnection = name

	updatedConfig, err := yaml.Marshal(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal config")
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, updatedConfig, 0644); err != nil {
		log.Error().Err(err).Msg("Failed to write config file")
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetCurrentConnection gets the current connection from the config file
func (c *Config) GetCurrentConnection() *MongoConfig {
	for _, connection := range c.Connections {
		if connection.Name == c.CurrentConnection {
			return &connection
		}
	}

	return nil
}

// AddConnection adds a MongoDB connection to the config file
func (c *Config) AddConnection(mongoConfig *MongoConfig) error {
	log.Info().Msgf("Adding connection: %s", mongoConfig.Name)
	if c.Connections == nil {
		c.Connections = []MongoConfig{}
	}
	for _, connection := range c.Connections {
		if connection.Name == mongoConfig.Name {
			return fmt.Errorf("connection with name %s already exists", mongoConfig.Name)
		}
	}
	c.Connections = append(c.Connections, *mongoConfig)

	updatedConfig, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, updatedConfig, 0644)
}

// AddConnectionFromUri adds a MongoDB connection to the config file
// using a URI
func (c *Config) AddConnectionFromUri(mongoConfig *MongoConfig) error {
	log.Info().Msgf("Adding connection from URI: %s", mongoConfig.GetSafeUri())
	host, port, db, err := ParseMongoDBURI(mongoConfig.Uri)
	if err != nil {
		return err
	}
	mongoConfig.Host = host
	intPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	mongoConfig.Port = intPort
	mongoConfig.Database = db
	return c.AddConnection(mongoConfig)
}

// DeleteConnection deletes a config from the config file by name
func (c *Config) DeleteConnection(name string) error {
	log.Info().Msgf("Deleting connection: %s", name)
	for i, connection := range c.Connections {
		if connection.Name == name {
			connection = MongoConfig{}
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
		}
	}

	updatedConfig, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, updatedConfig, 0644)
}

// GetUri returns the URI or builds it from the config
func (m *MongoConfig) GetUri() string {
	var uri string
	if m.Uri != "" {
		uri = m.Uri
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", m.Host, m.Port, m.Database)
		if m.Username != "" && m.Password != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", m.Username, m.Password, m.Host, m.Port, m.Database)
		}
	}

	return uri
}

// GetSafeUri returns the URI with the password replaced by asterisks
func (m *MongoConfig) GetSafeUri() string {
	uri := m.GetUri()
	return util.HidePasswordInUri(uri)
}

// GetOptions returns the options from the config file
// if they are not set, it returns the default values
func (c *MongoConfig) GetOptions() MongoOptions {
	boolPtr := true
	defaults := MongoOptions{
		AuthorizedDatabases:   &boolPtr,
		AuthorizedCollections: &boolPtr,
	}
	if c.Options.AuthorizedDatabases == nil {
		c.Options.AuthorizedDatabases = defaults.AuthorizedDatabases
	}
	if c.Options.AuthorizedCollections == nil {
		c.Options.AuthorizedCollections = defaults.AuthorizedCollections
	}
	return c.Options
}

func ParseMongoDBURI(uri string) (host, port, db string, err error) {
	if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
		return "", "", "", fmt.Errorf("invalid MongoDB URI prefix")
	}

	trimURI := strings.TrimPrefix(uri, "mongodb://")
	trimURI = strings.TrimPrefix(trimURI, "mongodb+srv://")

	splitURI := strings.Split(trimURI, "@")
	if len(splitURI) > 1 {
		trimURI = splitURI[1]
	} else {
		trimURI = splitURI[0]
	}

	if strings.Contains(trimURI, "?") {
		trimURI = strings.Split(trimURI, "?")[0]
	}

	splitDB := strings.Split(trimURI, "/")
	if len(splitDB) > 1 {
		db = splitDB[1]
		trimURI = splitDB[0]
	} else {
		db = ""
	}

	hosts := strings.Split(trimURI, ",")
	hostPort := strings.Split(hosts[0], ":")

	host = hostPort[0]
	if len(hostPort) > 1 {
		port = hostPort[1]
	} else {
		port = "27017"
	}

	return host, port, db, nil
}
