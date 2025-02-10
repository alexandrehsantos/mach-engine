package helpers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertJSONResponse is a helper function for testing JSON responses
func AssertJSONResponse(
	t *testing.T,
	rr *httptest.ResponseRecorder,
	expectedStatus int,
	expectedBody interface{},
) {
	t.Helper()

	// Assert status code
	assert.Equal(t, expectedStatus, rr.Code, "wrong status code")

	// Assert content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "wrong content type")

	// Assert body
	var actualBody interface{}
	err := json.NewDecoder(rr.Body).Decode(&actualBody)
	require.NoError(t, err, "failed to decode response body")
	assert.Equal(t, expectedBody, actualBody, "wrong response body")
}
