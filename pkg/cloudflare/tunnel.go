package cloudflare

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"proxy-service/pkg/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type TunnelConfig struct {
	ID                string
	Token             string
	CustomerID        string
	TargetURL         string
	HeartbeatInterval time.Duration
	RetryInterval     time.Duration
}

type TunnelState struct {
	Status      string
	LastHealthy time.Time
	Errors      int
	mutex       sync.RWMutex
}

type TunnelClient struct {
	config     TunnelConfig
	state      TunnelState
	connection *websocket.Conn
	httpClient *http.Client
	logger     *logger.Logger
	done       chan struct{}
}

func NewTunnelClient(config TunnelConfig, logger *logger.Logger) *TunnelClient {
	return &TunnelClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: true,
			},
		},
		logger: logger,
		done:   make(chan struct{}),
	}
}

func (c *TunnelClient) Start(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}

	go c.healthCheck()
	go c.handleMessages()

	return nil
}

func (c *TunnelClient) handleHealthCheckFailure(err error) {
	c.logger.Error("health check failed", zap.Error(err))
	c.state.mutex.Lock()
	c.state.Status = "error"
	c.state.Errors++
	c.state.mutex.Unlock()

	if c.state.Errors > 3 {
		c.reconnect()
	}
}

func (c *TunnelClient) processMessage(messageType int, message []byte) error {
	// Add your message processing logic here
	switch messageType {
	case websocket.TextMessage:
		c.logger.Debug("received text message", zap.ByteString("message", message))
	case websocket.BinaryMessage:
		c.logger.Debug("received binary message", zap.Int("size", len(message)))
	}
	return nil
}

func (c *TunnelClient) connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		HandshakeTimeout: 10 * time.Second,
	}

	// Connect to Cloudflare edge
	conn, _, err := dialer.DialContext(ctx, c.config.TargetURL, http.Header{
		"CF-Tunnel-Token": []string{c.config.Token},
		"CF-Tunnel-ID":    []string{c.config.ID},
	})
	if err != nil {
		return err
	}

	c.connection = conn
	c.updateState("connected", nil)
	return nil
}

func (c *TunnelClient) healthCheck() {
	ticker := time.NewTicker(c.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if err := c.sendHeartbeat(); err != nil {
				c.handleHealthCheckFailure(err)
			} else {
				c.updateState("healthy", nil)
			}
		}
	}
}

func (c *TunnelClient) sendHeartbeat() error {
	message := struct {
		Type      string    `json:"type"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Type:      "heartbeat",
		Timestamp: time.Now(),
	}

	return c.connection.WriteJSON(message)
}

func (c *TunnelClient) handleMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			messageType, message, err := c.connection.ReadMessage()
			if err != nil {
				c.handleConnectionError(err)
				return
			}

			if err := c.processMessage(messageType, message); err != nil {
				c.logger.Error("failed to process message", zap.Error(err))
			}
		}
	}
}

func (c *TunnelClient) handleConnectionError(err error) {
	c.updateState("error", err)
	c.reconnect()
}

func (c *TunnelClient) reconnect() {
	backoff := time.Second

	for {
		select {
		case <-c.done:
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := c.connect(ctx); err == nil {
				cancel()
				return
			}
			cancel()

			// Exponential backoff
			time.Sleep(backoff)
			backoff *= 2
			if backoff > time.Minute*5 {
				backoff = time.Minute * 5
			}
		}
	}
}

func (c *TunnelClient) updateState(status string, err error) {
	c.state.mutex.Lock()
	defer c.state.mutex.Unlock()

	c.state.Status = status
	if err == nil {
		c.state.LastHealthy = time.Now()
		c.state.Errors = 0
	} else {
		c.state.Errors++
	}
}

func (c *TunnelClient) ForwardRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	if c.state.Status != "healthy" {
		return nil, fmt.Errorf("tunnel not healthy")
	}

	// Clone and modify the request
	proxyReq := req.Clone(ctx)
	proxyReq.RequestURI = ""

	// Add tunnel headers
	proxyReq.Header.Set("CF-Tunnel-ID", c.config.ID)
	proxyReq.Header.Set("CF-Customer-ID", c.config.CustomerID)

	return c.httpClient.Do(proxyReq)
}

func (c *TunnelClient) Stop() error {
	close(c.done)
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}
