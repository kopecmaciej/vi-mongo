package config

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	pathToConfig        = "config.yaml"
	pathToDefaultConfig = "config.default.yaml"
)

type MongoConfig struct {
	Uri      string `yaml:"url"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type LogConfig struct {
	Path        string `yaml:"path"`
	Level       string `yaml:"level"`
	PrettyPrint bool   `yaml:"prettyPrint"`
}

type Config struct {
	Log                LogConfig     `yaml:"log"`
	Debug              bool          `yaml:"debug"`
	ShowConnectionPage bool          `yaml:"showConnectionPage"`
	CurrentConnection  string        `yaml:"currentConnection"`
	Connections        []MongoConfig `yaml:"connections"`
}

// LoadConfig loads the config file
// If the file does not exist, it will be created
// with the default settings
func LoadConfig() (*Config, error) {
	bytes, err := os.ReadFile(pathToConfig)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the config file with default settings
			bytes, err = os.ReadFile(pathToDefaultConfig)
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(pathToConfig, bytes, 0644)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	config := &Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// SaveMongoConfig saves the given MongoDB configuration to the config file
// If the file does not exist, it will be created
func SaveMongoConfig(config *MongoConfig) error {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(pathToConfig, bytes, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the config file with default settings
			err := os.WriteFile(pathToConfig, bytes, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// SetCurrentConnection sets the current connection in the config file
func (c *Config) SetCurrentConnection(name string) error {
	// If the user has set the alwaysShowConnectionPage setting to true,
	// we don't want to save the current connection
	c.CurrentConnection = name

	updatedConfig, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(pathToConfig, updatedConfig, 0644)
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
	c.Connections = append(c.Connections, *mongoConfig)

	updatedConfig, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(pathToConfig, updatedConfig, 0644)
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

	return os.WriteFile(pathToConfig, updatedConfig, 0644)
}

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

func loadAndUnmarshal() (*Config, error) {
	file, err := os.Open(pathToConfig)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.NewDecoder(file).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
