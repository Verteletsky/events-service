package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/godev/events-service/internal/config"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	logger   *zap.Logger
}

func NewMongoDB(ctx context.Context, logger *zap.Logger, cfg *config.MongoConfig) (*MongoDB, error) {
	client, err := mongo.Connect(ctx, cfg.Options)
	if err != nil {
		logger.Error("Failed to connect to MongoDB",
			zap.Error(err),
			zap.String("uri", cfg.URI))
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("Failed to ping MongoDB",
			zap.Error(err),
			zap.String("uri", cfg.URI))
		return nil, err
	}

	logger.Info("Successfully connected to MongoDB",
		zap.String("uri", cfg.URI))

	return &MongoDB{
		client:   client,
		database: client.Database(cfg.Database),
		logger:   logger,
	}, nil
}

func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func (m *MongoDB) GetDatabase() *mongo.Database {
	return m.database
}
