package app

import (
	"context"
	"proxy-service/internal/config"
	"proxy-service/internal/handler"
	"proxy-service/internal/service"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/cloudflare"
	"proxy-service/pkg/database"
	"proxy-service/pkg/logger"
	"proxy-service/pkg/metrics"
	"time"
)

type App struct {
	cfg    *config.Config
	server *Server
}

func NewApp(cfg *config.Config) (*App, error) {
	// Initialize MongoDB
	db, err := database.NewMongoDB(&cfg.MongoDB)
	if err != nil {
		return nil, err
	}

	// Initialize Redis
	redisClient, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		return nil, err
	}

	// Initialize metrics collector
	metricsCollector := metrics.NewMetricsCollector()

	log := logger.NewLogger()

	// Initialize Cloudflare tunnel client
	tunnelClient := cloudflare.NewTunnelClient(cloudflare.TunnelConfig{
		ID:                cfg.Cloudflare.TunnelID,
		Token:             cfg.Cloudflare.TunnelToken,
		TargetURL:         cfg.Proxy.TargetHost,
		HeartbeatInterval: 30 * time.Second,
		RetryInterval:     5 * time.Second,
	}, log)

	// Initialize services
	services := service.NewServices(service.Deps{
		Config:       cfg,
		DB:           db,
		Cache:        redisClient,
		Metrics:      metricsCollector,
		TunnelClient: tunnelClient,
	})

	// Initialize handlers
	handlers := handler.NewHandler(handler.Deps{
		Services: services,
		Config:   cfg,
	})

	// Initialize server
	server := NewServer(cfg, handlers)

	return &App{
		cfg:    cfg,
		server: server,
	}, nil
}

func (a *App) Start() error {
	return a.server.Start()
}

func (a *App) Stop(ctx context.Context) error {
	return a.server.Stop(ctx)
}
