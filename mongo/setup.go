package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func NewConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     27017,
		Username: "",
		Password: "",
		Database: "test",
	}
}

type Client struct {
	Client *mongo.Client
	config *Config
}

func NewClient(c *Config) *Client {
	return &Client{
		config: c,
	}
}

func (m *Client) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c := m.config
	var uri string
	uri = fmt.Sprintf("mongodb://%s:%d/%s", c.Host, c.Port, c.Database)
	if c.Username != "" && c.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
	}

	clientOptions := options.Client().ApplyURI(uri)

	var err error
	m.Client, err = mongo.Connect(ctx, clientOptions)

	return err
}

func (m *Client) Close(ctx context.Context) {
	m.Client.Disconnect(ctx)
}

func (m *Client) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx, nil)
}
