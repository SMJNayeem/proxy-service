package proxy

import (
	"net/http"
	"net/http/httputil"
	"proxy-service/internal/config"
	"proxy-service/pkg/metrics"
	"time"
)

type ProxyHandler struct {
	config  *config.Config
	metrics *metrics.MetricsCollector
}

func NewProxyHandler(config *config.Config, metrics *metrics.MetricsCollector) *ProxyHandler {
	return &ProxyHandler{
		config:  config,
		metrics: metrics,
	}
}

func (h *ProxyHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()
	customerID := request.Context().Value("customer_id").(string)

	director := func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = h.config.Proxy.TargetHost
		req.Host = h.config.Proxy.TargetHost
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			h.metrics.RecordError(customerID, "proxy_error")
			http.Error(w, err.Error(), http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(writer, request)

	duration := time.Since(start)
	h.metrics.RecordRequestDuration(customerID, request.URL.Path, request.Method, duration)
}
