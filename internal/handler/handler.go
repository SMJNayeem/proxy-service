package handler

import (
	"proxy-service/internal/config"
	"proxy-service/internal/service"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/metrics"
)

type Deps struct {
	Services *service.Services
	Config   *config.Config
	Cache    *cache.RedisCache
}

type Handler struct {
	Auth     *AuthHandler
	Proxy    *ProxyHandler
	Metrics  *MetricsHandler
	services *service.Services
	config   *config.Config
	cache    *cache.RedisCache
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		Auth:     NewAuthHandler(deps.Services.Auth),
		Proxy:    NewProxyHandler(deps.Services.Proxy, deps.Cache), // This is correct now
		Metrics:  NewMetricsHandler(deps.Services.Metrics),
		services: deps.Services,
		config:   deps.Config,
		cache:    deps.Cache,
	}
}

func (h *Handler) Config() *config.Config {
	return h.config
}

func (h *Handler) MetricsCollector() *metrics.MetricsCollector {
	if h.services != nil && h.services.Metrics != nil {
		return h.services.Metrics.GetCollector()
	}
	return nil
}

func (h *Handler) GetAuthService() *service.AuthService {
	return h.services.Auth
}

func (h *Handler) Cache() *cache.RedisCache {
	return h.cache
}
