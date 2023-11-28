package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
)

const configPath = "config.yaml"

type Config struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
	Database   string `yaml:"database"`
}

func NewConfig() *Config {
	return &Config{}
}

type Client struct {
	Client *mongo.Client
	Config *Config
}

func NewClient() *Client {
	c := NewConfig()
	return &Client{
		Config: c,
	}
}

func (m *Client) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := m.Config
	err := c.ReadYaml()

	var uri string
	uri = fmt.Sprintf("mongodb://%s:%d/%s", c.Host, c.Port, c.Database)
	if c.Username != "" && c.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
	}

	clientOptions := options.Client().ApplyURI(uri)
	m.Client, err = mongo.Connect(ctx, clientOptions)

	return err
}

func (m *Client) Close(ctx context.Context) {
	m.Client.Disconnect(ctx)
}

func (m *Client) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx, nil)
}

func (c *Config) ReadYaml() error {
	var yamlConfig struct {
		Mongo Config `yaml:"mongo"`
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = yaml.Unmarshal(file, &yamlConfig)
	if err != nil {
		return err
	}
	config := yamlConfig.Mongo

	c.Host = config.Host
	c.Port = config.Port
	c.Username = config.Username
	c.Password = config.Password
	c.Database = config.Database

	return nil
}
