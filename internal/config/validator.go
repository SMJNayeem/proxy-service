package config

import (
	"fmt"
	"strings"
)

func (c *Config) Validate() error {
	var errors []string

	// Validate Server config
	if c.Server.Port == "" {
		errors = append(errors, "server port is required")
	}
	if c.Server.TLSCertFile == "" {
		errors = append(errors, "TLS certificate file is required")
	}
	if c.Server.TLSKeyFile == "" {
		errors = append(errors, "TLS key file is required")
	}

	// Validate MongoDB config
	if c.MongoDB.URI == "" {
		errors = append(errors, "MongoDB URI is required")
	}
	if c.MongoDB.Database == "" {
		errors = append(errors, "MongoDB database name is required")
	}

	// Validate Redis config
	if c.Redis.Address == "" {
		errors = append(errors, "Redis address is required")
	}

	// Validate JWT config
	if c.JWT.Secret == "" {
		errors = append(errors, "JWT secret is required")
	}
	if c.JWT.ExpirationHours <= 0 {
		errors = append(errors, "JWT expiration hours must be positive")
	}

	// Validate Cloudflare config (moved from Cloudflare to c.Cloudflare)
	if c.Cloudflare.TunnelID == "" {
		errors = append(errors, "Cloudflare tunnel ID is required")
	}
	if c.Cloudflare.TunnelToken == "" {
		errors = append(errors, "Cloudflare tunnel token is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
