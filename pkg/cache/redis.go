package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"proxy-service/internal/config"
	"proxy-service/internal/models"
	"proxy-service/pkg/jwt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(cfg *config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) GetCustomer(ctx context.Context, key string) (*models.Customer, error) {
	data, err := c.client.Get(ctx, "customer:"+key).Result()
	if err != nil {
		return nil, err
	}

	var customer models.Customer
	if err := json.Unmarshal([]byte(data), &customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

func (c *RedisCache) SetCustomer(ctx context.Context, key string, customer *models.Customer, expiration time.Duration) error {
	data, err := json.Marshal(customer)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, "customer:"+key, data, expiration).Err()
}

func (c *RedisCache) GetProxyConfig(ctx context.Context, customerID string) (*models.ProxyConfig, error) {
	data, err := c.client.Get(ctx, "proxy_config:"+customerID).Result()
	if err != nil {
		return nil, err
	}

	var config models.ProxyConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *RedisCache) SetProxyConfig(ctx context.Context, customerID string, config *models.ProxyConfig, expiration time.Duration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, "proxy_config:"+customerID, data, expiration).Err()
}

func (c *RedisCache) GetTokenClaims(ctx context.Context, token string) (*jwt.Claims, error) {
	data, err := c.client.Get(ctx, "token:"+token).Result()
	if err != nil {
		return nil, err
	}

	var claims jwt.Claims
	if err := json.Unmarshal([]byte(data), &claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (c *RedisCache) SetTokenClaims(ctx context.Context, token string, claims *jwt.Claims, expiration time.Duration) error {
	data, err := json.Marshal(claims)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "token:"+token, data, expiration).Err()
}

func (c *RedisCache) GetRoutePermission(ctx context.Context, customerID, route string) (bool, bool) {
	key := fmt.Sprintf("route_permission:%s:%s", customerID, route)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return false, false
	}
	return val == "true", true
}

func (c *RedisCache) SetRoutePermission(ctx context.Context, customerID, route string, allowed bool, expiration time.Duration) error {
	key := fmt.Sprintf("route_permission:%s:%s", customerID, route)
	value := "false"
	if allowed {
		value = "true"
	}
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisCache) GetAgentConfig(ctx context.Context, key string) (*models.AgentConfig, error) {
	data, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var config models.AgentConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *RedisCache) SetAgentConfig(ctx context.Context, key string, config *models.AgentConfig, expiration time.Duration) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return c.Set(ctx, key, string(data), expiration)
}
