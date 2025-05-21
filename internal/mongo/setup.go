package mongo

import (
	"context"
	"time"

	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/util"

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
	if m.Config.Password != "" && config.EncryptionKey != "" {
		password, err := util.DecryptPassword(m.Config.Password, config.EncryptionKey)
		if err != nil {
			return err
		}
		uri = util.RestorePasswordInUri(uri, password)

	}
	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}

	m.Client = client

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
