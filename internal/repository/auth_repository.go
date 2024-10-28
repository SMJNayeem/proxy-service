package repository

import (
	"context"
	"proxy-service/internal/models"
	"time"

	"proxy-service/pkg/cache"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthRepository struct {
	db    *mongo.Database
	cache *cache.RedisCache
}

func NewAuthRepository(db *mongo.Database, cache *cache.RedisCache) *AuthRepository {
	return &AuthRepository{
		db:    db,
		cache: cache,
	}
}

func (r *AuthRepository) GetCustomerByAPIKey(ctx context.Context, apiKey string) (*models.Customer, error) {
	// Try cache first
	if customer, err := r.cache.GetCustomer(ctx, apiKey); err == nil {
		return customer, nil
	}

	// Query database
	var customer models.Customer
	err := r.db.Collection("customers").FindOne(ctx, bson.M{"api_key": apiKey}).Decode(&customer)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.SetCustomer(ctx, apiKey, &customer, time.Hour)
	return &customer, nil
}

func (r *AuthRepository) GetCustomer(ctx context.Context, customerID string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Collection("customers").FindOne(ctx, bson.M{"_id": customerID}).Decode(&customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}
