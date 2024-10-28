package service

import (
	"context"
	"fmt"
	"net/http"
	"proxy-service/internal/models"
	"proxy-service/internal/service/agent"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/metrics"
	"sync"
	"time"
)

type ProxyService struct {
	agentManager *agent.AgentManager
	cache        *cache.RedisCache
	metrics      *metrics.MetricsCollector
	routingTable map[string]string // customerID -> agentID
	routingMutex sync.RWMutex
}

type ProxyRequest struct {
	Method     string
	Path       string
	Headers    http.Header
	Body       []byte
	CustomerID string
}

type ProxyResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func NewProxyService(
	agentManager *agent.AgentManager,
	cache *cache.RedisCache,
	metrics *metrics.MetricsCollector,
) *ProxyService {
	service := &ProxyService{
		agentManager: agentManager,
		cache:        cache,
		metrics:      metrics,
		routingTable: make(map[string]string),
	}

	// Start routing table maintenance
	go service.maintainRoutingTable()

	return service
}

func (s *ProxyService) ForwardRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	startTime := time.Now()
	customerID := ctx.Value("customer_id").(string)

	// Get proxy configuration
	_, err := s.getProxyConfig(ctx, customerID)
	if err != nil {
		s.metrics.RecordError(customerID, "config_error")
		return nil, fmt.Errorf("failed to get proxy config: %w", err)
	}

	// Get agent ID for the customer
	agentID, err := s.getAgentForCustomer(customerID)
	if err != nil {
		s.metrics.RecordError(customerID, "routing_error")
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Create proxy request
	proxyReq := &agent.ProxyRequest{
		Method:     req.Method,
		Path:       req.URL.Path,
		Headers:    req.Header,
		CustomerID: customerID,
	}

	// Forward request through agent
	response, err := s.agentManager.RouteRequest(ctx, agentID, proxyReq)
	if err != nil {
		s.metrics.RecordError(customerID, "forward_error")
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}

	// Record metrics
	s.metrics.RecordRequestDuration(customerID, req.URL.Path, req.Method, time.Since(startTime))

	// Convert agent.ProxyResponse to ProxyResponse
	localResponse := &ProxyResponse{
		StatusCode: response.StatusCode,
		Headers:    response.Headers,
		Body:       response.Body,
	}

	return s.createHTTPResponse(localResponse), nil
}

func (s *ProxyService) getProxyConfig(ctx context.Context, customerID string) (*models.ProxyConfig, error) {
	// Try cache first
	if config, err := s.cache.GetProxyConfig(ctx, customerID); err == nil {
		return config, nil
	}

	// Config not in cache, return error
	return nil, fmt.Errorf("proxy configuration not found")
}

func (s *ProxyService) getAgentForCustomer(customerID string) (string, error) {
	s.routingMutex.RLock()
	agentID, exists := s.routingTable[customerID]
	s.routingMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("no agent found for customer")
	}

	return agentID, nil
}

func (s *ProxyService) maintainRoutingTable() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateRoutingTable()
		}
	}
}

func (s *ProxyService) updateRoutingTable() {
	// Get all active agents
	agents := s.agentManager.GetActiveAgents()

	s.routingMutex.Lock()
	defer s.routingMutex.Unlock()

	// Clear current routing table
	s.routingTable = make(map[string]string)

	// Update routing table based on active agents
	for _, agent := range agents {
		s.routingTable[agent.CustomerID] = agent.AgentID
	}
}

func (s *ProxyService) createHTTPResponse(proxyResp *ProxyResponse) *http.Response {
	return &http.Response{
		StatusCode: proxyResp.StatusCode,
		Header:     proxyResp.Headers,
		Body:       http.NoBody,
	}
}
