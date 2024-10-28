package database

import (
	"context"
	"fmt"
	"proxy-service/internal/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDB(cfg *config.MongoDBConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(uint64(cfg.PoolSize)).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(time.Duration(cfg.Timeout) * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Ping database to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return &MongoDB{
		client: client,
		db:     client.Database(cfg.Database),
	}, nil
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

func (m *MongoDB) Database() *mongo.Database {
	return m.db
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
