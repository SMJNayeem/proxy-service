package agent

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"proxy-service/internal/config"
	"proxy-service/internal/models"
	"proxy-service/internal/service/agent"
	"proxy-service/pkg/cache"
	"proxy-service/pkg/logger"
	"proxy-service/pkg/metrics"
	"proxy-service/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Message types for WebSocket communication
const (
	MessageTypeProxyRequest  = "proxy_request"
	MessageTypeProxyResponse = "proxy_response"
	MessageTypeHeartbeat     = "heartbeat"
	MessageTypeConfig        = "config_update"
	MessageTypeMetrics       = "metrics_update"
	MessageTypeError         = "error"

	AgentTokenExpiry  = 24 * time.Hour
	TokenCachePrefix  = "agent_token:"
	HeartbeatInterval = 30 * time.Second
)

type WSMessage struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type AgentHandler struct {
	agentManager *agent.AgentManager
	metrics      *metrics.MetricsCollector
	upgrader     websocket.Upgrader
	authService  AuthService
	cache        *cache.RedisCache
	config       *config.Config
	logger       *logger.Logger
	validator    *validator.RequestValidator
}

type AgentSecurityConfig struct {
	TLSConfig      *tls.Config
	IPWhitelist    []string
	MaxRequestSize int64
	RequestTimeout time.Duration
}

func NewAgentHandler(
	agentManager *agent.AgentManager,
	authService AuthService,
	cache *cache.RedisCache,
	config *config.Config,
	metrics *metrics.MetricsCollector,
	logger *logger.Logger,
) *AgentHandler {
	h := &AgentHandler{
		agentManager: agentManager,
		authService:  authService,
		cache:        cache,
		config:       config,
		metrics:      metrics,
		logger:       logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 1. Get agent and customer IDs from headers
				agentID := r.Header.Get("X-Agent-ID")
				customerID := r.Header.Get("X-Customer-ID")

				// 2. Validate basic requirements
				if agentID == "" || customerID == "" {
					return false
				}

				// 3. Check origin against allowed domains if configured
				if config.Agent.AllowedOrigins != nil && len(config.Agent.AllowedOrigins) > 0 {
					origin := r.Header.Get("Origin")
					if origin == "" {
						return false
					}

					// Check if origin is in allowed list
					for _, allowed := range config.Agent.AllowedOrigins {
						if origin == allowed {
							return true
						}
					}
					return false
				}

				// 4. If no specific origins are configured, allow only authenticated agents
				return true // Authentication is handled in HandleConnection
			},
			HandshakeTimeout: 10 * time.Second,
		},
	}
	h.validator = validator.NewRequestValidator(&config.Agent, cache, logger)
	return h
}

// HandleConnection is the main WebSocket endpoint for agent communication
func (h *AgentHandler) HandleConnection(c *gin.Context) {
	// 1. Validate agent credentials
	agentID := c.GetHeader("X-Agent-ID")
	customerID := c.GetHeader("X-Customer-ID")
	token := c.GetHeader("X-Agent-Token")

	if err := h.validateAgentCredentials(agentID, customerID, token); err != nil {
		h.logger.Error("Agent authentication failed", "error", err, "agent_id", agentID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent credentials"})
		return
	}

	// 2. Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", "error", err, "agent_id", agentID)
		return
	}

	// 3. Initialize agent connection
	agentConn := &models.AgentConnection{
		ID:         agentID,
		CustomerID: customerID,
		Conn:       conn,
		StartTime:  time.Now(),
		LastPing:   time.Now(),
	}

	// 4. Register agent with manager
	if err := h.agentManager.RegisterAgent(c.Request.Context(), agentID, customerID, conn); err != nil {
		h.logger.Error("Agent registration failed", "error", err, "agent_id", agentID)
		conn.Close()
		return
	}

	// 5. Record connection metric
	h.metrics.RecordAgentConnection(customerID)

	// 6. Start message handling routines
	go h.handleMessages(agentConn)
	go h.monitorHeartbeat(agentConn)

	h.logger.Info("Agent connected successfully", "agent_id", agentID, "customer_id", customerID)
}

func (h *AgentHandler) handleHeartbeat(agent *models.AgentConnection) {
	agent.LastPing = time.Now()
	h.metrics.RecordAgentHeartbeat(agent.CustomerID, agent.ID)

	// Check for config updates
	if config := h.agentManager.GetConfigUpdate(agent.ID); config != nil {
		h.sendMessage(agent, MessageTypeConfig, config)
	}
}

func (h *AgentHandler) handleProxyRequest(agent *models.AgentConnection, msg *WSMessage) {
	var proxyReq models.ProxyRequest
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", msg.Payload)), &proxyReq); err != nil {
		h.logger.Error("Failed to decode proxy request", "error", err, "agent_id", agent.ID)
		h.sendError(agent, msg.RequestID, "invalid request format")
		return
	}

	// Process proxy request through manager
	response, err := h.agentManager.HandleProxyRequest(context.Background(), &proxyReq)
	if err != nil {
		h.logger.Error("Proxy request failed", "error", err, "agent_id", agent.ID)
		h.sendError(agent, msg.RequestID, err.Error())
		return
	}

	// Send response back to agent
	h.sendMessage(agent, MessageTypeProxyResponse, response)
}

func (h *AgentHandler) handleMetricsUpdate(agent *models.AgentConnection, msg *WSMessage) {
	var metrics models.AgentMetrics
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", msg.Payload)), &metrics); err != nil {
		h.logger.Error("Failed to decode metrics", "error", err, "agent_id", agent.ID)
		return
	}

	h.metrics.UpdateAgentMetrics(agent.CustomerID, agent.ID, &metrics)
}

func (h *AgentHandler) monitorHeartbeat(agent *models.AgentConnection) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Since(agent.LastPing) > HeartbeatInterval*3 {
				h.logger.Warn("Agent heartbeat timeout",
					zap.String("agent_id", agent.ID),
					zap.Duration("last_ping", time.Since(agent.LastPing)))
				agent.Conn.Close()
				return
			}
		}
	}
}

func (h *AgentHandler) handleMessages(agent *models.AgentConnection) {
	defer func() {
		h.handleDisconnection(agent)
	}()

	for {
		messageType, message, err := agent.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket read error",
					zap.Error(err),
					zap.String("agent_id", agent.ID))
			}
			return
		}

		if messageType == websocket.TextMessage {
			var wsMessage WSMessage
			if err := json.Unmarshal(message, &wsMessage); err != nil {
				h.logger.Error("Failed to parse message",
					zap.Error(err),
					zap.String("agent_id", agent.ID))
				continue
			}

			switch wsMessage.Type {
			case MessageTypeHeartbeat:
				h.handleHeartbeat(agent)
			case MessageTypeProxyRequest:
				h.handleProxyRequest(agent, &wsMessage)
			case MessageTypeMetrics:
				h.handleMetricsUpdate(agent, &wsMessage)
			default:
				h.logger.Warn("Unknown message type",
					zap.String("message_type", wsMessage.Type),
					zap.String("agent_id", agent.ID))
			}
		}
	}
}

func (h *AgentHandler) handleDisconnection(agent *models.AgentConnection) {
	h.agentManager.DeregisterAgent(context.Background(), agent.ID, agent.CustomerID)
	h.metrics.RecordAgentDisconnection(agent.CustomerID)
	agent.Conn.Close()
	h.logger.Info("Agent disconnected", "agent_id", agent.ID)
}

func (h *AgentHandler) sendMessage(agent *models.AgentConnection, messageType string, payload interface{}) {
	msg := WSMessage{
		Type:      messageType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	if err := agent.Conn.WriteJSON(msg); err != nil {
		h.logger.Error("Failed to send message", "error", err, "agent_id", agent.ID)
	}
}

func (h *AgentHandler) sendError(agent *models.AgentConnection, requestID, errorMsg string) {
	h.sendMessage(agent, MessageTypeError, gin.H{
		"request_id": requestID,
		"error":      errorMsg,
	})
}

// func NewAgentHandler(agentManager *agent.AgentManager, metrics *metrics.MetricsCollector) *AgentHandler {
// 	return &AgentHandler{
// 		agentManager: agentManager,
// 		metrics:      metrics,
// 		upgrader: websocket.Upgrader{
// 			ReadBufferSize:  1024,
// 			WriteBufferSize: 1024,
// 			CheckOrigin: func(r *http.Request) bool {
// 				// Implement proper origin checking
// 				return true
// 			},
// 			HandshakeTimeout: 10 * time.Second,
// 		},
// 	}
// }

// HandleConnection handles new agent websocket connections
// func (h *AgentHandler) HandleConnection(c *gin.Context) {
// 	// Validate agent credentials
// 	agentID := c.GetHeader("X-Agent-ID")
// 	customerID := c.GetHeader("X-Customer-ID")
// 	token := c.GetHeader("X-Agent-Token")

// 	if err := h.validateAgentCredentials(agentID, customerID, token); err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent credentials"})
// 		return
// 	}

// 	// Upgrade connection to WebSocket
// 	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket upgrade failed"})
// 		return
// 	}

// 	// Register agent
// 	if err := h.agentManager.RegisterAgent(c.Request.Context(), agentID, customerID, conn); err != nil {
// 		conn.Close()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "agent registration failed"})
// 		return
// 	}

// 	// Record metrics
// 	h.metrics.RecordAgentConnection(customerID)
// }

func (h *AgentHandler) validateAgentCredentials(agentID, customerID, token string) error {
	ctx := context.Background()

	// 1. Basic validation
	if agentID == "" || customerID == "" || token == "" {
		return fmt.Errorf("missing required credentials")
	}

	// 2. Check cache first for rate limiting and recently validated tokens
	cacheKey := fmt.Sprintf("%s%s_%s", TokenCachePrefix, customerID, agentID)
	if cachedToken, err := h.cache.Get(ctx, cacheKey); err == nil {
		if cachedToken == token {
			return nil // Token recently validated
		}
	}

	// 3. Get customer details
	customer, err := h.authService.GetCustomer(ctx, customerID)
	if err != nil {
		return fmt.Errorf("invalid customer: %w", err)
	}

	// 4. Verify customer status
	if customer.Status != "active" {
		return fmt.Errorf("customer account is not active")
	}

	// 5. Get agent details
	agent, err := h.authService.GetAgent(ctx, agentID)
	if err != nil {
		return fmt.Errorf("invalid agent: %w", err)
	}

	// 6. Verify agent belongs to customer
	if agent.CustomerID != customerID {
		return fmt.Errorf("agent does not belong to customer")
	}

	// 7. Verify agent status
	if agent.Status != "active" {
		return fmt.Errorf("agent is not active")
	}

	// 8. Validate token format and expiry
	if err := h.validateTokenFormat(token); err != nil {
		return fmt.Errorf("invalid token format: %w", err)
	}

	// 9. Verify token signature
	expectedToken := h.generateAgentToken(agent, customer.APIKey)
	if !h.compareTokens(token, expectedToken) {
		return fmt.Errorf("invalid token signature")
	}

	// 10. Cache the validated token
	err = h.cache.Set(ctx, cacheKey, token, 15*time.Minute)
	if err != nil {
		h.logger.Error("failed to cache token", "error", err)
		// Don't return error here as validation was successful
	}

	return nil
}

func (h *AgentHandler) validateTokenFormat(token string) error {
	// Token format: timestamp.signature
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid token format")
	}

	// Verify timestamp
	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}

	// Check token expiry
	if time.Unix(timestamp, 0).Add(AgentTokenExpiry).Before(time.Now()) {
		return fmt.Errorf("token expired")
	}

	return nil
}

func (h *AgentHandler) generateAgentToken(agent *models.Agent, apiKey string) string {
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("%s:%s:%d", agent.ID, agent.CustomerID, timestamp)

	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%d.%s", timestamp, signature)
}

func (h *AgentHandler) compareTokens(providedToken, expectedToken string) bool {
	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(providedToken), []byte(expectedToken))
}

// AuthService interface for dependency injection
type AuthService interface {
	GetCustomer(ctx context.Context, customerID string) (*models.Customer, error)
	GetAgent(ctx context.Context, agentID string) (*models.Agent, error)
}

// HandleAgentHealth handles agent health check endpoints
func (h *AgentHandler) HandleAgentHealth(c *gin.Context) {
	agentID := c.GetHeader("X-Agent-ID")
	if status := h.agentManager.GetAgentStatus(agentID); status == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (h *AgentHandler) HandleAgentDeregistration(c *gin.Context) {
	agentID := c.GetHeader("X-Agent-ID")
	customerID := c.GetHeader("X-Customer-ID")

	if err := h.agentManager.DeregisterAgent(c.Request.Context(), agentID, customerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// Record metric for agent deregistration
	h.metrics.RecordAgentDisconnection(customerID)

	c.JSON(http.StatusOK, gin.H{
		"status":   "deregistered",
		"agent_id": agentID,
	})
}

// Add these methods to AgentHandler
func (h *AgentHandler) HandleAgentMetrics(c *gin.Context) {
	agentID := c.GetHeader("X-Agent-ID")
	customerID := c.GetHeader("X-Customer-ID")

	metrics, err := h.agentManager.GetAgentMetrics(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// Record metrics
	h.metrics.UpdateAgentUptime(customerID, agentID, metrics.Uptime)
	h.metrics.UpdateAgentMemoryUsage(customerID, agentID, metrics.MemoryUsage)
	h.metrics.UpdateAgentCPUUsage(customerID, agentID, metrics.CPUUsage)

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"agent_id": agentID,
		"metrics":  metrics,
	})
}

func (h *AgentHandler) HandleAgentRegistration(c *gin.Context) {
	// Implementation for agent registration
	var agent models.Agent
	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Add your registration logic
	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}
