package cache

import (
	"context"
	"proxy-service/internal/models"
	"proxy-service/pkg/jwt"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	GetCustomer(ctx context.Context, key string) (*models.Customer, error)
	SetCustomer(ctx context.Context, key string, customer *models.Customer, expiration time.Duration) error
	GetProxyConfig(ctx context.Context, key string) (*models.ProxyConfig, error)
	SetProxyConfig(ctx context.Context, key string, config *models.ProxyConfig, expiration time.Duration) error
	GetTokenClaims(ctx context.Context, token string) (*jwt.Claims, error)
	SetTokenClaims(ctx context.Context, token string, claims *jwt.Claims, expiration time.Duration) error
	GetRoutePermission(ctx context.Context, customerID, route string) (bool, bool)
	SetRoutePermission(ctx context.Context, customerID, route string, allowed bool, expiration time.Duration) error
}
