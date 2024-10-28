package handler

import (
	"io"
	"net/http"
	"proxy-service/internal/service"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	proxyService *service.ProxyService
	logger       *logger.Logger
	cache        *cache.RedisCache
}

func NewProxyHandler(service *service.ProxyService, cache *cache.RedisCache) *ProxyHandler {
	return &ProxyHandler{
		proxyService: service,
		logger:       logger.NewLogger(),
		cache:        cache,
	}
}

func (h *ProxyHandler) HandleRequest(c *gin.Context) {
	customerID := c.GetString("customer_id")

	// Get proxy configuration
	_, err := h.cache.GetProxyConfig(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get proxy configuration"})
		return
	}

	// Forward the request
	resp, err := h.proxyService.ForwardRequest(c.Request.Context(), c.Request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "proxy request failed"})
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set status code
	c.Status(resp.StatusCode)

	// Copy body
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to copy response"})
		return
	}
}
