package integration

// import (
// 	"context"
// 	"net/http"
// 	"proxy-service/internal/config"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestProxyIntegration(t *testing.T) {
// 	ctx := context.Background()
// 	cfg, err := config.Load("../../config/config.test.yaml")
// 	assert.NoError(t, err)

// 	// Setup test server
// 	ts := setupTestServer()
// 	defer ts.Close()

// 	// Test cases
// 	tests := []struct {
// 		name           string
// 		path           string
// 		method         string
// 		expectedStatus int
// 	}{
// 		{
// 			name:           "Valid Request",
// 			path:           "/api/test",
// 			method:         "GET",
// 			expectedStatus: http.StatusOK,
// 		},
// 		{
// 			name:           "Invalid Path",
// 			path:           "/invalid",
// 			method:         "GET",
// 			expectedStatus: http.StatusNotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Make request
// 			resp, err := makeRequest(tt.method, ts.URL+tt.path)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
// 		})
// 	}
// }
