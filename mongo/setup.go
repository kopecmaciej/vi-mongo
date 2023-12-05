package mongo

import (
	"context"
	"fmt"
	"github.com/kopecmaciej/mongui/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const configPath = "config.yaml"

type Client struct {
	Client *mongo.Client
	Config *config.MongoConfig
}

func NewClient(config *config.MongoConfig) *Client {
	return &Client{
		Config: config,
	}
}

func (m *Client) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := config.LoadAppConfig()
	if err != nil {
		return err
	}
	c := config.Mongo

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
