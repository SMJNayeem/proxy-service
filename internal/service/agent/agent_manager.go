package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"proxy-service/internal/models"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/logger"
	"proxy-service/pkg/metrics"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type AgentConnection struct {
	AgentID    string
	CustomerID string
	Connection *websocket.Conn
	Status     string
	LastPing   time.Time
	mutex      sync.RWMutex
}

type AgentManager struct {
	connections map[string]*AgentConnection
	metrics     *metrics.MetricsCollector
	cache       *cache.Cache
	mutex       sync.RWMutex
	logger      *logger.Logger
}

type AgentMetrics struct {
	Uptime      float64   `json:"uptime"`
	MemoryUsage float64   `json:"memory_usage"`
	CPUUsage    float64   `json:"cpu_usage"`
	LastUpdated time.Time `json:"last_updated"`
}

const (
	configCacheKey = "agent_config:%s"
	configTTL      = 5 * time.Minute
)

func NewAgentManager(metrics *metrics.MetricsCollector, cache *cache.Cache) *AgentManager {
	manager := &AgentManager{
		connections: make(map[string]*AgentConnection),
		metrics:     metrics,
		cache:       cache,
	}

	// Start cleanup routine
	go manager.cleanupInactiveAgents()

	return manager
}

type ProxyRequest struct {
	Method     string
	Path       string
	Headers    map[string][]string
	Body       []byte
	CustomerID string
}

type ProxyResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

func (am *AgentManager) RegisterAgent(ctx context.Context, agentID, customerID string, conn *websocket.Conn) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Check if agent already exists
	if existing, exists := am.connections[agentID]; exists {
		existing.Connection.Close()
		delete(am.connections, agentID)
	}

	// Create new agent connection
	agent := &AgentConnection{
		AgentID:    agentID,
		CustomerID: customerID,
		Connection: conn,
		Status:     "connected",
		LastPing:   time.Now(),
	}

	// Store connection
	am.connections[agentID] = agent

	// Record metric
	am.metrics.RecordAgentConnection(customerID)

	// Start monitoring routine
	go am.monitorAgent(agent)

	return nil
}

func (am *AgentManager) GetActiveAgents() []*AgentConnection {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agents := make([]*AgentConnection, 0, len(am.connections))
	for _, agent := range am.connections {
		if agent.Status == "connected" {
			agents = append(agents, agent)
		}
	}

	return agents
}

func (am *AgentManager) RouteRequest(ctx context.Context, agentID string, request *ProxyRequest) (*ProxyResponse, error) {
	am.mutex.RLock()
	agent, exists := am.connections[agentID]
	am.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent not found")
	}

	return agent.sendRequest(request)
}

func (ac *AgentConnection) sendRequest(request *ProxyRequest) (*ProxyResponse, error) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	// Send request through websocket
	if err := ac.Connection.WriteJSON(request); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	response := &ProxyResponse{}
	if err := ac.Connection.ReadJSON(response); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return response, nil
}

func (am *AgentManager) cleanupInactiveAgents() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.mutex.Lock()
			for id, agent := range am.connections {
				if time.Since(agent.LastPing) > 5*time.Minute {
					agent.Connection.Close()
					delete(am.connections, id)
					am.metrics.RecordAgentDisconnection(agent.CustomerID)
				}
			}
			am.mutex.Unlock()
		}
	}
}

func (am *AgentManager) GetAgentStatus(agentID string) string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	if agent, exists := am.connections[agentID]; exists {
		return agent.Status
	}
	return ""
}

func (am *AgentManager) checkAgentHealth(agent *AgentConnection) error {
	agent.mutex.Lock()
	defer agent.mutex.Unlock()

	if time.Since(agent.LastPing) > 2*time.Minute {
		return fmt.Errorf("agent timeout")
	}

	// Send ping message
	if err := agent.Connection.WriteControl(
		websocket.PingMessage,
		[]byte{},
		time.Now().Add(10*time.Second),
	); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	return nil
}

func (am *AgentManager) handleAgentDisconnection(agentID string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if agent, exists := am.connections[agentID]; exists {
		// Record metric before removing
		am.metrics.RecordAgentDisconnection(agent.CustomerID)

		// Close connection
		agent.Connection.Close()

		// Remove from connections map
		delete(am.connections, agentID)
	}
}

func (am *AgentManager) monitorAgent(agent *AgentConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := am.checkAgentHealth(agent); err != nil {
				am.handleAgentDisconnection(agent.AgentID)
				return
			}
		}
	}
}

func (am *AgentManager) DeregisterAgent(ctx context.Context, agentID, customerID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if agent, exists := am.connections[agentID]; exists {
		if agent.CustomerID != customerID {
			return fmt.Errorf("unauthorized deregistration attempt")
		}

		// Close the connection
		if err := agent.Connection.Close(); err != nil {
			return fmt.Errorf("error closing connection: %w", err)
		}

		// Remove from connections map
		delete(am.connections, agentID)
		return nil
	}

	return fmt.Errorf("agent not found")
}

func (am *AgentManager) GetAgentMetrics(ctx context.Context, agentID string) (*AgentMetrics, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agent, exists := am.connections[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found")
	}

	// In a real implementation, you would collect these metrics from the agent
	metrics := &AgentMetrics{
		Uptime:      time.Since(agent.LastPing).Seconds(),
		MemoryUsage: 0, // Implement actual memory usage collection
		CPUUsage:    0, // Implement actual CPU usage collection
		LastUpdated: time.Now(),
	}

	return metrics, nil
}

func (am *AgentManager) GetConfigUpdate(agentID string) *models.AgentConfig {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agent, exists := am.connections[agentID]
	if !exists {
		return nil
	}

	// Return any pending configuration updates
	return am.getAgentConfig(agent.CustomerID)
}

func (am *AgentManager) HandleProxyRequest(ctx context.Context, req *models.ProxyRequest) (*models.ProxyResponse, error) {
	am.mutex.RLock()
	agent, exists := am.connections[req.AgentID]
	am.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent not found")
	}

	// Create proxy request
	proxyReq := &ProxyRequest{
		Method:     req.Method,
		Path:       req.Path,
		Headers:    convertHeaders(req.Headers), // Convert map[string]string to map[string][]string
		Body:       req.Body,
		CustomerID: agent.CustomerID,
	}

	// Use the connection to send the request
	proxyResp, err := agent.sendRequest(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to proxy request: %w", err)
	}

	// Convert to models.ProxyResponse
	response := &models.ProxyResponse{
		RequestID:  req.RequestID,
		StatusCode: proxyResp.StatusCode,
		Headers:    convertHeadersBack(proxyResp.Headers),
		Body:       proxyResp.Body,
	}

	am.logger.Info("Handling proxy request",
		zap.String("agent_id", req.AgentID),
		zap.String("path", req.Path),
		zap.String("method", req.Method))

	return response, nil
}

func (am *AgentManager) getAgentConfig(customerID string) *models.AgentConfig {
	ctx := context.Background()

	// Try to get from cache first
	cacheKey := fmt.Sprintf(configCacheKey, customerID)

	// Using the generic Get method since we're working with the Cache interface
	configData, err := (*am.cache).Get(ctx, cacheKey)
	if err == nil {
		var config models.AgentConfig
		if err := json.Unmarshal([]byte(configData), &config); err == nil {
			return &config
		}
	}

	// If not in cache or error occurred, create default configuration
	config := &models.AgentConfig{
		CustomerID:        customerID,
		MaxConnections:    100,
		RequestTimeout:    30 * time.Second,
		RetryAttempts:     3,
		RetryDelay:        5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		Features: models.AgentFeatures{
			EnableCompression: true,
			EnableCaching:     true,
			EnableMetrics:     true,
		},
		Security: models.SecurityConfig{
			EnableTLS:      true,
			MinTLSVersion:  "1.2",
			AllowedOrigins: []string{"*"},
			RateLimit: models.RateLimitConfig{
				Enabled:    true,
				Requests:   1000,
				TimeWindow: time.Minute,
			},
		},
		Routes: []models.RouteConfig{
			{
				Path:         "/**",
				Methods:      []string{"GET", "POST", "PUT", "DELETE"},
				RateLimit:    1000,
				Timeout:      30 * time.Second,
				CacheEnabled: true,
				CacheTTL:     5 * time.Minute,
			},
		},
		Monitoring: models.MonitoringConfig{
			MetricsInterval: 30 * time.Second,
			LogLevel:        "info",
			EnableTracing:   true,
			SamplingRate:    0.1,
		},
	}

	// Store in cache for future use
	if configData, err := json.Marshal(config); err == nil {
		err = (*am.cache).Set(ctx, cacheKey, string(configData), configTTL)
		if err != nil {
			am.logger.Error("Failed to cache agent config",
				zap.Error(err),
				zap.String("customer_id", customerID))
		}
	}

	return config
}

// Add this helper function
func convertHeaders(headers map[string]string) map[string][]string {
	result := make(map[string][]string)
	for k, v := range headers {
		result[k] = []string{v}
	}
	return result
}

// Add this helper function at the bottom of the file
func convertHeadersBack(headers map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
