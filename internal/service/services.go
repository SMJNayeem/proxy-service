package service

import (
	"proxy-service/internal/config"
	"proxy-service/internal/repository"
	"proxy-service/internal/service/agent"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/cloudflare"
	"proxy-service/pkg/database"
	"proxy-service/pkg/metrics"
)

type Services struct {
	Auth    *AuthService
	Proxy   *ProxyService
	Metrics *MetricsService
}

type Deps struct {
	Config       *config.Config
	DB           *database.MongoDB
	Cache        *cache.RedisCache
	Metrics      *metrics.MetricsCollector
	TunnelClient *cloudflare.TunnelClient
	cache        *cache.Cache
}

func NewServices(deps Deps) *Services {
	db := deps.DB.Database()

	authRepo := repository.NewAuthRepository(db, deps.Cache)
	// proxyRepo := repository.NewProxyRepository(db, deps.Cache)
	metricsRepo := repository.NewMetricsRepository(db)

	agentManager := agent.NewAgentManager(deps.Metrics, deps.cache)

	authService, _ := NewAuthService(authRepo, deps.Cache, deps.Config, deps.Metrics)
	proxyService := NewProxyService(agentManager, deps.Cache, deps.Metrics)
	metricsService := NewMetricsService(metricsRepo, deps.Metrics)

	return &Services{
		Auth:    authService,
		Proxy:   proxyService,
		Metrics: metricsService,
	}
}
