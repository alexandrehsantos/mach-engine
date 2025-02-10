package mocks

import (
	"net/http"
)

// MockResponseWriter is a mock implementation of http.ResponseWriter
type MockResponseWriter struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		Headers: make(http.Header),
	}
}

func (m *MockResponseWriter) Header() http.Header {
	return m.Headers
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	m.Body = b
	return len(b), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}
