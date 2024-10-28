package repository

import (
	"context"
	"proxy-service/internal/models"
	"proxy-service/pkg/cache"

	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProxyRepository struct {
	db    *mongo.Database
	cache *cache.RedisCache
}

func NewProxyRepository(db *mongo.Database, cache *cache.RedisCache) *ProxyRepository {
	return &ProxyRepository{
		db:    db,
		cache: cache,
	}
}

func (r *ProxyRepository) GetConfig(ctx context.Context, customerID string) (*models.ProxyConfig, error) {
	// Try cache first
	if config, err := r.cache.GetProxyConfig(ctx, customerID); err == nil {
		return config, nil
	}

	// Query database
	var config models.ProxyConfig
	err := r.db.Collection("proxy_configs").FindOne(ctx, bson.M{"customer_id": customerID}).Decode(&config)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.SetProxyConfig(ctx, customerID, &config, time.Hour)
	return &config, nil
}
