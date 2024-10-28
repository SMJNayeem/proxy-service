package middleware

import (
	"net/http"
	"proxy-service/internal/handler"
	"proxy-service/internal/service"
	"proxy-service/pkg/metrics"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *service.AuthService
	metrics     *metrics.MetricsCollector
}

func NewAuthMiddleware(authService *service.AuthService, metrics *metrics.MetricsCollector) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		metrics:     metrics,
	}
}

func (m *AuthMiddleware) ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Get token from header
		token := c.GetHeader("Authorization")
		if token == "" {
			m.metrics.RecordAuthFailure("missing_token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
				"code":  "AUTH_HEADER_MISSING",
			})
			return
		}

		// Remove Bearer prefix
		token = strings.TrimPrefix(token, "Bearer ")

		// Validate token
		claims, err := m.authService.VerifyToken(c.Request.Context(), token)
		if err != nil {
			m.metrics.RecordAuthFailure("invalid_token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
				"code":  "INVALID_TOKEN",
			})
			return
		}

		// Check token expiration
		if time.Now().After(claims.ExpiresAt.Time) {
			m.metrics.RecordAuthFailure("token_expired")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token expired",
				"code":  "TOKEN_EXPIRED",
			})
			return
		}

		// Check if route is allowed for this token
		requestPath := c.Request.URL.Path
		if !m.authService.IsRouteAllowed(c.Request.Context(), claims.CustomerID, requestPath) {
			m.metrics.RecordAuthFailure("route_unauthorized")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "route not authorized",
				"code":  "ROUTE_UNAUTHORIZED",
			})
			return
		}

		// Set claims in context
		c.Set("customer_id", claims.CustomerID)
		c.Set("allowed_routes", claims.AllowedRoutes)

		// Record successful auth
		m.metrics.RecordAuthSuccess("token_validated")
		m.metrics.RecordAuthDuration(time.Since(startTime))

		c.Next()
	}
}

func Auth(h *handler.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
				"code":  "AUTH_HEADER_MISSING",
			})
			return
		}

		// Remove Bearer prefix if present
		token = strings.TrimPrefix(token, "Bearer ")

		// Use the exposed GetAuthService method
		authService := h.GetAuthService()

		// Verify token
		claims, err := authService.VerifyToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
				"code":  "INVALID_TOKEN",
			})
			return
		}

		// Set customer ID and routes in context
		c.Set("customer_id", claims.CustomerID)
		c.Set("allowed_routes", claims.AllowedRoutes)

		c.Next()
	}
}
