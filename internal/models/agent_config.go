package models

import "time"

type AgentConfig struct {
	CustomerID        string            `json:"customer_id"`
	MaxConnections    int               `json:"max_connections"`
	RequestTimeout    time.Duration     `json:"request_timeout"`
	RetryAttempts     int               `json:"retry_attempts"`
	RetryDelay        time.Duration     `json:"retry_delay"`
	HeartbeatInterval time.Duration     `json:"heartbeat_interval"`
	Features          AgentFeatures     `json:"features"`
	Security          SecurityConfig    `json:"security"`
	Routes            []RouteConfig     `json:"routes"`
	Monitoring        MonitoringConfig  `json:"monitoring"`
	Settings          map[string]string `json:"settings"`
	LastUpdated       time.Time         `json:"last_updated"`
}

type AgentFeatures struct {
	EnableCompression bool `json:"enable_compression"`
	EnableCaching     bool `json:"enable_caching"`
	EnableMetrics     bool `json:"enable_metrics"`
}

type SecurityConfig struct {
	EnableTLS      bool            `json:"enable_tls"`
	MinTLSVersion  string          `json:"min_tls_version"`
	AllowedOrigins []string        `json:"allowed_origins"`
	RateLimit      RateLimitConfig `json:"rate_limit"`
}

type RateLimitConfig struct {
	Enabled    bool          `json:"enabled"`
	Requests   int           `json:"requests"`
	TimeWindow time.Duration `json:"time_window"`
}

type RouteConfig struct {
	Path         string        `json:"path"`
	Methods      []string      `json:"methods"`
	RateLimit    int           `json:"rate_limit"`
	Timeout      time.Duration `json:"timeout"`
	CacheEnabled bool          `json:"cache_enabled"`
	CacheTTL     time.Duration `json:"cache_ttl"`
}

type MonitoringConfig struct {
	MetricsInterval time.Duration `json:"metrics_interval"`
	LogLevel        string        `json:"log_level"`
	EnableTracing   bool          `json:"enable_tracing"`
	SamplingRate    float64       `json:"sampling_rate"`
}
