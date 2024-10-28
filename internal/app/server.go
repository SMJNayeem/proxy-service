package app

import (
	"context"
	"crypto/tls"
	"net/http"
	"proxy-service/internal/config"
	"proxy-service/internal/handler"
	"proxy-service/internal/middleware"
	"proxy-service/pkg/logger"
	"proxy-service/pkg/validator"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	handler    *handler.Handler
}

// func NewServer(cfg *config.Config, handler *handler.Handler) *Server {
// 	// Initialize router
// 	router := gin.New()

// 	// Add middlewares
// 	router.Use(gin.Recovery())
// 	router.Use(middleware.Logger())
// 	router.Use(middleware.Metrics(handler.MetricsCollector()))

// 	// Setup routes
// 	api := router.Group("/api/v1")
// 	{
// 		// Auth routes
// 		auth := api.Group("/auth")
// 		{
// 			auth.POST("/token", handler.Auth.GenerateToken)
// 			auth.POST("/verify", handler.Auth.VerifyToken)
// 		}

// 		// Protected routes
// 		protected := api.Group("")
// 		protected.Use(middleware.Auth(handler.Auth))
// 		{
// 			// Proxy routes
// 			protected.Any("/*path", handler.Proxy.HandleRequest)

// 			// Metrics routes
// 			protected.GET("/metrics", handler.Metrics.GetMetrics)
// 		}
// 	}

// 	// Configure TLS
// 	tlsConfig := &tls.Config{
// 		MinVersion:               tls.VersionTLS12,
// 		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
// 		PreferServerCipherSuites: true,
// 	}

// 	// Create HTTP server
// 	return &Server{
// 		httpServer: &http.Server{
// 			Addr:              ":" + cfg.Server.Port,
// 			Handler:           router,
// 			TLSConfig:         tlsConfig,
// 			ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
// 			WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
// 			IdleTimeout:       time.Duration(cfg.Server.IdleTimeout) * time.Second,
// 			ReadHeaderTimeout: 20 * time.Second,
// 		},
// 		handler: handler,
// 	}
// }

func NewServer(cfg *config.Config, handler *handler.Handler) *Server {
	// Initialize router
	router := gin.New()

	// Get cache from handler
	cache := handler.Cache()

	// Initialize logger
	log := logger.NewLogger()

	// Create request validator
	validator := validator.NewRequestValidator(&cfg.Agent, cache, log)

	// Create resilience middleware
	resilience := middleware.NewResilienceMiddleware(validator)

	// Add middlewares
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.Metrics(handler.MetricsCollector()))
	router.Use(resilience.Handle())

	// Setup routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/token", handler.Auth.GenerateToken)
			auth.POST("/verify", handler.Auth.VerifyToken)
		}

		// Protected routes
		protected := api.Group("")
		// Pass the entire handler instead of just the Auth handler
		protected.Use(middleware.Auth(handler))
		{
			// Proxy routes
			protected.Any("/*path", handler.Proxy.HandleRequest)

			// Metrics routes
			protected.GET("/metrics", handler.Metrics.GetMetrics)
		}
	}

	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.Server.Port,
			Handler:           router,
			TLSConfig:         tlsConfig,
			ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
			WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
			IdleTimeout:       time.Duration(cfg.Server.IdleTimeout) * time.Second,
			ReadHeaderTimeout: 20 * time.Second,
		},
		handler: handler,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServeTLS(
		s.handler.Config().Server.TLSCertFile,
		s.handler.Config().Server.TLSKeyFile,
	)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
