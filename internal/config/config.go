package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	MongoDB    MongoDBConfig    `mapstructure:"mongodb"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	Proxy      ProxyConfig      `mapstructure:"proxy"`
	Cloudflare CloudflareConfig `mapstructure:"cloudflare"`
	Agent      AgentConfig      `mapstructure:"agent"`
}

type CloudflareConfig struct {
	TunnelID          string        `mapstructure:"tunnel_id"`
	TunnelToken       string        `mapstructure:"tunnel_token"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	RetryInterval     time.Duration `mapstructure:"retry_interval"`
	MaxRetries        int           `mapstructure:"max_retries"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
	HandshakeTimeout  time.Duration `mapstructure:"handshake_timeout"`
}

type AgentConfig struct {
	MaxConnections    int            `mapstructure:"max_connections"`
	ConnectionTimeout time.Duration  `mapstructure:"connection_timeout"`
	HeartbeatInterval time.Duration  `mapstructure:"heartbeat_interval"`
	MaxRequestTimeout time.Duration  `mapstructure:"max_request_timeout"`
	MaxRetries        int            `mapstructure:"max_retries"`
	RetryInterval     time.Duration  `mapstructure:"retry_interval"`
	BufferSize        int            `mapstructure:"buffer_size"`
	AllowedOrigins    []string       `mapstructure:"allowed_origins"`
	Security          SecurityConfig `mapstructure:"security"`
	MaxRequestSize    int64          `mapstructure:"max_request_size"`
}

type SecurityConfig struct {
	AllowedOrigins []string        `mapstructure:"allowed_origins"`
	RateLimit      RateLimitConfig `mapstructure:"rate_limit"`
}

type RateLimitConfig struct {
	Requests   int           `mapstructure:"requests"`
	TimeWindow time.Duration `mapstructure:"time_window"`
}

// Add new ProxyConfig struct
type ProxyConfig struct {
	TargetHost string `mapstructure:"target_host"`
}

type ServerConfig struct {
	Port         string `mapstructure:"port"`
	TLSCertFile  string `mapstructure:"tls_cert_file"`
	TLSKeyFile   string `mapstructure:"tls_key_file"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
	PoolSize int    `mapstructure:"pool_size"`
	Timeout  int    `mapstructure:"timeout"`
}

type RedisConfig struct {
	Address      string `mapstructure:"address"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	MaxRetries   int    `mapstructure:"max_retries"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
