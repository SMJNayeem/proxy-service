package models

import "errors"

type ProxyConfig struct {
	ID            string            `bson:"_id" json:"id"`
	CustomerID    string            `bson:"customer_id" json:"customer_id"`
	TargetURL     string            `bson:"target_url" json:"target_url"`
	Headers       map[string]string `bson:"headers" json:"headers"`
	RateLimit     int               `bson:"rate_limit" json:"rate_limit"`
	Timeout       int               `bson:"timeout" json:"timeout"`
	RetryCount    int               `bson:"retry_count" json:"retry_count"`
	CacheEnabled  bool              `bson:"cache_enabled" json:"cache_enabled"`
	TunnelEnabled bool              `bson:"tunnel_enabled" json:"tunnel_enabled"`
}

type ProxyRoute struct {
	Path         string `bson:"path" json:"path"`
	Method       string `bson:"method" json:"method"`
	RateLimit    int    `bson:"rate_limit" json:"rate_limit"`
	CacheEnabled bool   `bson:"cache_enabled" json:"cache_enabled"`
}

type ProxyRequest struct {
	RequestID  string            `json:"request_id"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	CustomerID string            `json:"customer_id"`
	AgentID    string            `json:"agent_id"`
}

type ProxyResponse struct {
	RequestID  string            `json:"request_id"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Error      string            `json:"error,omitempty"`
}

func (req *ProxyRequest) Validate() error {
	if req.CustomerID == "" {
		return errors.New("customer ID is required")
	}
	if req.Path == "" {
		return errors.New("path is required")
	}
	// Add more validation rules
	return nil
}
