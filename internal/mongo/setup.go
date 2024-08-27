package mongo

import (
	"context"
	"time"

	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/rs/zerolog/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.Config.Timeout)*time.Second)
	defer cancel()

	uri := m.Config.GetUri()
	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, opts)
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

func (m *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.Config.Timeout)*time.Second)
	defer cancel()
	return m.Client.Ping(ctx, nil)
}
