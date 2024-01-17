package mongo

import (
	"context"
	"time"

	"github.com/kopecmaciej/mongui/config"
	"github.com/rs/zerolog/log"

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

	uri := m.Config.GetUri()
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	m.Client = client

  log.Info().Msgf("Connected to %s", uri)

	return nil
}

func (m *Client) Close(ctx context.Context) {
	m.Client.Disconnect(ctx)
}

func (m *Client) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx, nil)
}
