package service

import (
	"context"
	"errors"
	"path"
	"proxy-service/internal/config"
	"proxy-service/internal/models"
	"proxy-service/internal/repository"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/jwt"
	"proxy-service/pkg/metrics"
	"strings"
	"time"
)

var (
	ErrCustomerInactive = errors.New("customer is inactive")
	ErrInvalidAPIKey    = errors.New("invalid API key")
)

type AuthService struct {
	repo    *repository.AuthRepository
	cache   *cache.RedisCache
	config  *config.Config
	metrics *metrics.MetricsCollector
	jwtMgr  *jwt.JWTManager
}

func NewAuthService(repo *repository.AuthRepository, cache *cache.RedisCache, config *config.Config, metrics *metrics.MetricsCollector) (*AuthService, error) {
	// Initialize JWT manager
	jwtMgr, err := jwt.NewJWTManager(config.JWT.Secret, "", "proxy-service") // Empty string for public key as we're using HMAC
	if err != nil {
		return nil, err
	}

	return &AuthService{
		repo:    repo,
		cache:   cache,
		config:  config,
		metrics: metrics,
		jwtMgr:  jwtMgr,
	}, nil
}

// Add the GenerateToken method
func (s *AuthService) GenerateToken(ctx context.Context, apiKey string) (*models.AuthResponse, error) {
	// Get customer by API key
	customer, err := s.repo.GetCustomerByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, ErrInvalidAPIKey
	}

	// Check if customer is active
	if customer.Status != "active" {
		return nil, ErrCustomerInactive
	}

	// Generate token duration
	duration := time.Duration(s.config.JWT.ExpirationHours) * time.Hour

	// Generate JWT token
	token, err := s.jwtMgr.GenerateToken(customer.ID, customer.AllowedRoutes, duration)
	if err != nil {
		return nil, err
	}

	// Create response
	response := &models.AuthResponse{
		Token:     token,
		ExpiresIn: s.config.JWT.ExpirationHours * 3600, // Convert hours to seconds
	}

	return response, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, token string) (*jwt.Claims, error) {
	// Try cache first
	if claims, err := s.cache.GetTokenClaims(ctx, token); err == nil {
		return claims, nil
	}

	// Validate token
	claims, err := s.jwtMgr.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Verify customer exists and is active
	customer, err := s.repo.GetCustomer(ctx, claims.CustomerID)
	if err != nil {
		return nil, err
	}

	if customer.Status != "active" {
		return nil, ErrCustomerInactive
	}

	// Cache validated token
	s.cache.SetTokenClaims(ctx, token, claims, time.Until(time.Now().Add(time.Duration(s.config.JWT.ExpirationHours)*time.Hour)))

	return claims, nil
}

func (s *AuthService) IsRouteAllowed(ctx context.Context, customerID, route string) bool {
	// Check route permissions in cache first
	if allowed, exists := s.cache.GetRoutePermission(ctx, customerID, route); exists {
		return allowed
	}

	// Get customer routes from database
	customer, err := s.repo.GetCustomer(context.Background(), customerID)
	if err != nil {
		return false
	}

	// Check if route is allowed
	allowed := false
	for _, allowedRoute := range customer.AllowedRoutes {
		if matchRoute(allowedRoute, route) {
			allowed = true
			break
		}
	}

	// Cache the result
	s.cache.SetRoutePermission(ctx, customerID, route, allowed, time.Hour)

	return allowed
}

func matchRoute(pattern, route string) bool {
	// Convert pattern and route to path format
	pattern = path.Clean("/" + pattern)
	route = path.Clean("/" + route)

	// Split into segments
	patternParts := strings.Split(pattern, "/")
	routeParts := strings.Split(route, "/")

	// If lengths don't match and pattern doesn't end with /**
	if len(patternParts) != len(routeParts) && !strings.HasSuffix(pattern, "/**") {
		return false
	}

	// Compare each segment
	for i, patternPart := range patternParts {
		if i >= len(routeParts) {
			return false
		}

		// Handle wildcards
		switch patternPart {
		case "*":
			continue
		case "**":
			return true
		default:
			if patternPart != routeParts[i] {
				return false
			}
		}
	}

	return len(routeParts) == len(patternParts)
}
