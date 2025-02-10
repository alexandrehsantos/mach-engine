package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "company.com/matchengine/internal/handler/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// HealthCheckTestSuite defines the test suite
type HealthCheckTestSuite struct {
	suite.Suite
	handler http.HandlerFunc
}

// SetupTest prepares the test suite
func (s *HealthCheckTestSuite) SetupTest() {
	s.handler = handler.HealthCheck
}

// TestHealthCheckSuite runs the test suite
func TestHealthCheckSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckTestSuite))
}

// TestHealthCheck tests the health check handler
func (s *HealthCheckTestSuite) TestHealthCheck() {
	tests := []struct {
		name            string
		method          string
		path            string
		headers         map[string]string
		expectedStatus  int
		expectedBody    map[string]string
		expectedHeaders map[string]string
	}{
		{
			name:           "successful health check with GET",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "ok"},
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			path:           "/health",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   map[string]string{"error": "method not allowed"},
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name:   "with request ID header",
			method: http.MethodGet,
			path:   "/health",
			headers: map[string]string{
				"X-Request-ID": "test-request-id",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "ok"},
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
				"X-Request-ID": "test-request-id",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Setup
			req := httptest.NewRequest(tt.method, tt.path, nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			rr := httptest.NewRecorder()

			// Execute
			s.handler(rr, req)

			// Assert
			s.assertResponse(rr, tt.expectedStatus, tt.expectedBody, tt.expectedHeaders)
		})
	}
}

// Helper method for response assertions
func (s *HealthCheckTestSuite) assertResponse(
	rr *httptest.ResponseRecorder,
	expectedStatus int,
	expectedBody map[string]string,
	expectedHeaders map[string]string,
) {
	// Status code assertion
	assert.Equal(s.T(), expectedStatus, rr.Code, "wrong status code")

	// Body assertion
	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(s.T(), err, "failed to decode response body")
	assert.Equal(s.T(), expectedBody, response, "unexpected response body")

	// Headers assertion
	for key, value := range expectedHeaders {
		assert.Equal(s.T(), value, rr.Header().Get(key), "wrong header value for %s", key)
	}
}

// BenchmarkHealthCheck benchmarks the health check handler
func BenchmarkHealthCheck(b *testing.B) {
	handler := handler.HealthCheck
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		handler(rr, req)
	}
}
