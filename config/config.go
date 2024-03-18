package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	Timeout  int    `yaml:"timeout"`
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

	return os.WriteFile(pathToConfig, updatedConfig, 0644)
}

// AddConnectionFromUri adds a MongoDB connection to the config file
// using a URI
func (c *Config) AddConnectionFromUri(mongoConfig *MongoConfig) error {
	log.Info().Msgf("Adding connection from URI: %s", mongoConfig.Uri)
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

	splitDB := strings.Split(trimURI, "/")
	if len(splitDB) > 1 {
		db = splitDB[1]
	}
	if strings.Contains(db, "?") {
		db = strings.Split(db, "?")[0]
	}
	if strings.HasPrefix(uri, "mongodb+srv://") {
		host = trimURI
		port = "Default SRV"
		return host, port, db, nil
	}

	hostPort := strings.Split(trimURI, "/")[0]
	hostPortSplit := strings.Split(hostPort, ":")
	host = hostPortSplit[0]
	if len(hostPortSplit) > 1 {
		port = hostPortSplit[1]
	} else {
		port = "27017"
	}
	return host, port, db, nil
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
