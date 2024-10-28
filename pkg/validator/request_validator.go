package validator

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"proxy-service/internal/config"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/logger"
)

// RateLimiter implements a simple token bucket algorithm
type RateLimiter struct {
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
	mutex      sync.Mutex
}

func NewRateLimiter(capacity float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate

	if rl.tokens > rl.capacity {
		rl.tokens = rl.capacity
	}

	if rl.tokens < 1 {
		return false
	}

	rl.tokens--
	rl.lastRefill = now
	return true
}

type RequestValidator struct {
	config       *config.AgentConfig
	cache        *cache.RedisCache
	logger       *logger.Logger
	rateLimiters map[string]*RateLimiter
	mutex        sync.RWMutex
}

func NewRequestValidator(config *config.AgentConfig, cache *cache.RedisCache, logger *logger.Logger) *RequestValidator {
	return &RequestValidator{
		config:       config,
		cache:        cache,
		logger:       logger,
		rateLimiters: make(map[string]*RateLimiter),
	}
}

func (rv *RequestValidator) ValidateRequest(req *http.Request) error {
	// 1. IP Whitelist Check
	if err := rv.validateIP(req); err != nil {
		return fmt.Errorf("IP validation failed: %w", err)
	}

	// 2. Request Size Validation
	if err := rv.validateRequestSize(req); err != nil {
		return fmt.Errorf("size validation failed: %w", err)
	}

	// 3. Rate Limiting
	if err := rv.validateRateLimit(req); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// 4. Additional Security Headers Check
	if err := rv.validateSecurityHeaders(req); err != nil {
		return fmt.Errorf("security headers validation failed: %w", err)
	}

	return nil
}

func (rv *RequestValidator) validateIP(req *http.Request) error {
	// Get client IP
	clientIP := getClientIP(req)

	// If no whitelist is configured, allow all
	if len(rv.config.Security.AllowedOrigins) == 0 {
		return nil
	}

	// Check if IP is in whitelist
	for _, allowed := range rv.config.Security.AllowedOrigins {
		if allowed == "*" {
			return nil
		}

		// Handle CIDR notation
		if strings.Contains(allowed, "/") {
			_, ipNet, err := net.ParseCIDR(allowed)
			if err != nil {
				continue
			}
			if ipNet.Contains(net.ParseIP(clientIP)) {
				return nil
			}
		}

		// Direct IP comparison
		if allowed == clientIP {
			return nil
		}
	}

	return fmt.Errorf("IP %s not in whitelist", clientIP)
}

func (rv *RequestValidator) validateRequestSize(req *http.Request) error {
	// Check Content-Length header
	if req.ContentLength > rv.config.MaxRequestSize {
		return fmt.Errorf("request size %d exceeds maximum allowed size %d",
			req.ContentLength, rv.config.MaxRequestSize)
	}

	return nil
}

func (rv *RequestValidator) validateRateLimit(req *http.Request) error {
	// Get agent ID from header
	agentID := req.Header.Get("X-Agent-ID")
	if agentID == "" {
		return fmt.Errorf("missing agent ID")
	}

	// Get or create rate limiter for this agent
	rv.mutex.Lock()
	limiter, exists := rv.rateLimiters[agentID]
	if !exists {
		limiter = NewRateLimiter(
			float64(rv.config.Security.RateLimit.Requests),
			float64(rv.config.Security.RateLimit.Requests)/
				rv.config.Security.RateLimit.TimeWindow.Seconds(),
		)
		rv.rateLimiters[agentID] = limiter
	}
	rv.mutex.Unlock()

	// Check rate limit
	if !limiter.Allow() {
		return fmt.Errorf("rate limit exceeded for agent %s", agentID)
	}

	return nil
}

func (rv *RequestValidator) validateSecurityHeaders(req *http.Request) error {
	// Validate required headers
	requiredHeaders := map[string]bool{
		"X-Agent-ID":    true,
		"X-Customer-ID": true,
		"X-Agent-Token": true,
		"Content-Type":  true,
	}

	for header := range requiredHeaders {
		if req.Header.Get(header) == "" {
			return fmt.Errorf("missing required header: %s", header)
		}
	}

	// Validate Content-Type
	contentType := req.Header.Get("Content-Type")
	allowedContentTypes := map[string]bool{
		"application/json":                  true,
		"application/x-www-form-urlencoded": true,
		"multipart/form-data":               true,
	}

	if !allowedContentTypes[contentType] {
		return fmt.Errorf("unsupported Content-Type: %s", contentType)
	}

	return nil
}

// Helper function to get client IP from request
func getClientIP(req *http.Request) string {
	// Check X-Real-IP header
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Check X-Forwarded-For header
	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Get the first IP in the list
		if i := strings.Index(forwardedFor, ","); i != -1 {
			return strings.TrimSpace(forwardedFor[:i])
		}
		return strings.TrimSpace(forwardedFor)
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}
